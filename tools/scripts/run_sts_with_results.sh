#!/bin/bash

# STS (Strategic Test Suite) Benchmark Runner with Results Recording
# Enhanced script that records STS benchmark results to history

# Default values
DEPTH=999
TIMEOUT=5
MAX_POSITIONS=10  # Per file - total will be MAX_POSITIONS * number of files
VERBOSE=true
EPD_DIR="testdata"
RESULTS_FILE="sts_history.md"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--depth)
            DEPTH="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -m|--max)
            MAX_POSITIONS="$2"
            shift 2
            ;;
        --dir)
            EPD_DIR="$2"
            shift 2
            ;;
        -o|--output)
            RESULTS_FILE="$2"
            shift 2
            ;;
        -q|--quiet)
            VERBOSE=false
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -d, --depth N        Search depth (default: $DEPTH)"
            echo "  -t, --timeout N      Timeout per position in seconds (default: ${TIMEOUT}s)"
            echo "  -m, --max N          Max positions per file (default: $MAX_POSITIONS)"
            echo "  --dir PATH           EPD directory path (default: $EPD_DIR)"
            echo "  -o, --output PATH    Results file path (default: $RESULTS_FILE)"
            echo "  -q, --quiet          Disable verbose output"
            echo "  -h, --help           Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                   # Run with defaults: 10 pos/file × 6 files = 60 total"
            echo "  $0 -t 3 -m 5         # 3s timeout, 5 pos/file × 6 files = 30 total"
            echo "  $0 -t 10 -m 20       # 10s timeout, 20 pos/file × 6 files = 120 total"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Interactive timeout prompt if not provided via command line
if [[ ! " $* " =~ " -t " ]] && [[ ! " $* " =~ " --timeout " ]]; then
    echo "STS Benchmark Runner with Results Recording"
    echo "==========================================="
    echo ""
    read -p "Enter timeout per position in seconds (default: ${TIMEOUT}): " USER_TIMEOUT
    if [[ -n "$USER_TIMEOUT" && "$USER_TIMEOUT" =~ ^[0-9]+$ ]]; then
        TIMEOUT=$USER_TIMEOUT
    fi
    echo ""
else
    echo "STS Benchmark Runner with Results Recording"
    echo "==========================================="
fi
echo "Configuration:"
echo "  Search Depth: $DEPTH"
echo "  Timeout: ${TIMEOUT}s per position"
echo "  Max Positions: $MAX_POSITIONS per file"
echo "  EPD Directory: $EPD_DIR"
echo "  Results File: $RESULTS_FILE"
echo "  Verbose: $VERBOSE"
echo ""

# Check if EPD directory exists and find STS files
if [ ! -d "$EPD_DIR" ]; then
    echo "Error: EPD directory '$EPD_DIR' not found!"
    exit 1
fi

# Find all STS*.epd files
EPD_FILES=($(find "$EPD_DIR" -name "STS*.epd" | sort))
if [ ${#EPD_FILES[@]} -eq 0 ]; then
    echo "Error: No STS*.epd files found in '$EPD_DIR'!"
    exit 1
fi

echo "Found ${#EPD_FILES[@]} STS files:"
for file in "${EPD_FILES[@]}"; do
    echo "  $(basename "$file")"
done
TOTAL_POSITIONS=$((MAX_POSITIONS * ${#EPD_FILES[@]}))
echo "Total positions to test: $TOTAL_POSITIONS (${MAX_POSITIONS} per file)"
echo ""

echo "Building STS benchmark..."
go build -o bin/sts ./cmd/sts

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Running STS benchmark..."
echo "======================="

# Prepare verbose flag
VERBOSE_FLAG=""
if [ "$VERBOSE" = true ]; then
    VERBOSE_FLAG="-verbose"
fi

# Run the benchmark and capture output while displaying in real-time
TIMESTAMP=$(date '+%Y-%m-%d %H:%M')
START_TIME=$(date +%s)

# Create temporary file to capture all output
TEMP_OUTPUT=$(mktemp)
ALL_OUTPUT=""

# Initialize aggregate results
TOTAL_SCORE=0
TOTAL_MAX_SCORE=0
TOTAL_POSITIONS_TESTED=0
TOTAL_CORRECT_MOVES=0
TOTAL_DEPTH_SUM=0
TOTAL_NODES=0

# Run benchmark for each EPD file
for EPD_FILE in "${EPD_FILES[@]}"; do
    echo "Processing $(basename "$EPD_FILE")..."
    echo "=================================="
    
    # Run command with tee to show real-time output and capture it
    ./bin/sts -depth "$DEPTH" -timeout "$TIMEOUT" -max "$MAX_POSITIONS" -file "$EPD_FILE" $VERBOSE_FLAG 2>&1 | tee "$TEMP_OUTPUT"
    EXIT_CODE=${PIPESTATUS[0]}
    
    if [ $EXIT_CODE -ne 0 ]; then
        echo "STS benchmark failed for $(basename "$EPD_FILE") with exit code $EXIT_CODE"
        rm "$TEMP_OUTPUT"
        exit $EXIT_CODE
    fi
    
    # Read and append output for parsing
    FILE_OUTPUT=$(cat "$TEMP_OUTPUT")
    ALL_OUTPUT="$ALL_OUTPUT$FILE_OUTPUT"$'\n'
    
    # Parse individual file results
    FILE_SCORE=$(echo "$FILE_OUTPUT" | grep -E "Total score:" | sed -E 's/.*Total score: ([0-9]+)\/([0-9]+).*/\1/')
    FILE_MAX_SCORE=$(echo "$FILE_OUTPUT" | grep -E "Total score:" | sed -E 's/.*Total score: ([0-9]+)\/([0-9]+).*/\2/')
    FILE_POSITIONS=$(echo "$FILE_OUTPUT" | grep -E "Positions tested:" | sed -E 's/.*Positions tested: ([0-9]+).*/\1/')
    FILE_CORRECT=$(echo "$FILE_OUTPUT" | grep -E "Correct moves \(10 points\):" | sed -E 's/.*Correct moves \(10 points\): ([0-9]+)\/([0-9]+).*/\1/')
    
    # Aggregate results
    TOTAL_SCORE=$((TOTAL_SCORE + FILE_SCORE))
    TOTAL_MAX_SCORE=$((TOTAL_MAX_SCORE + FILE_MAX_SCORE))
    TOTAL_POSITIONS_TESTED=$((TOTAL_POSITIONS_TESTED + FILE_POSITIONS))
    TOTAL_CORRECT_MOVES=$((TOTAL_CORRECT_MOVES + FILE_CORRECT))
    
    echo ""
done

END_TIME=$(date +%s)
TOTAL_TIME=$((END_TIME - START_TIME))

# Use aggregated output for final parsing
OUTPUT="$ALL_OUTPUT"
rm "$TEMP_OUTPUT"

if [ $EXIT_CODE -ne 0 ]; then
    echo "STS benchmark failed with exit code $EXIT_CODE"
    exit $EXIT_CODE
fi

# Use aggregated results
SCORE=$TOTAL_SCORE
MAX_SCORE=$TOTAL_MAX_SCORE
POSITIONS_TESTED=$TOTAL_POSITIONS_TESTED
CORRECT_MOVES=$TOTAL_CORRECT_MOVES

# Calculate aggregated metrics using bash arithmetic
SCORE_PERCENT_INT=$((SCORE * 100 / MAX_SCORE))
SCORE_PERCENT="${SCORE_PERCENT_INT}"

AVG_TIME_SECONDS=$((TOTAL_TIME / POSITIONS_TESTED))
AVG_TIME="${AVG_TIME_SECONDS}s"

# Parse performance metrics from final aggregated output (last occurrence)
AVG_DEPTH=$(echo "$ALL_OUTPUT" | grep -E "Average depth:" | tail -1 | sed -E 's/.*Average depth: ([0-9.]+).*/\1/')
NPS_RAW=$(echo "$ALL_OUTPUT" | grep -E "Nodes per second:" | tail -1 | sed -E 's/.*Nodes per second: ([0-9.]+[KM]?).*/\1/')

# Default values if parsing fails
if [ -z "$AVG_DEPTH" ]; then
    AVG_DEPTH="0.0"
fi
if [ -z "$NPS_RAW" ]; then
    NPS_RAW="0"
fi

# Calculate STS rating based on percentage (simplified approximation)
if [ $SCORE_PERCENT_INT -ge 90 ]; then
    STS_RATING=$((3400 + (SCORE_PERCENT_INT - 90) * 20 / 10))
elif [ $SCORE_PERCENT_INT -ge 80 ]; then
    STS_RATING=$((3200 + (SCORE_PERCENT_INT - 80) * 20 / 10))
elif [ $SCORE_PERCENT_INT -ge 70 ]; then
    STS_RATING=$((3000 + (SCORE_PERCENT_INT - 70) * 20 / 10))
elif [ $SCORE_PERCENT_INT -ge 60 ]; then
    STS_RATING=$((2700 + (SCORE_PERCENT_INT - 60) * 30 / 10))
elif [ $SCORE_PERCENT_INT -ge 50 ]; then
    STS_RATING=$((2400 + (SCORE_PERCENT_INT - 50) * 30 / 10))
else
    STS_RATING=$((2000 + SCORE_PERCENT_INT * 8 / 10))
fi

# Get current git commit for tracking version
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Create EPD files description
EPD_DESCRIPTION="STS1-${#EPD_FILES[@]} (${#EPD_FILES[@]} files)"

# Create or update results file
if [ ! -f "$RESULTS_FILE" ]; then
    echo "# STS Benchmark Results History" > "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
    echo "This file tracks the STS (Strategic Test Suite) performance of ChessEngine over time." >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
    echo "## Results Summary" >> "$RESULTS_FILE"
    echo "" >> "$RESULTS_FILE"
    echo "| Date | Commit | EPD File | Positions | Score | Max | Percent | STS Rating | Correct | Depth | Timeout | Avg Time | Total Time | NPS | Avg Depth | Notes |" >> "$RESULTS_FILE"
    echo "|------|--------|----------|-----------|-------|-----|---------|------------|---------|-------|---------|----------|------------|-----|-----------|-------|" >> "$RESULTS_FILE"
fi

# Prepare notes field (can be customized)
NOTES="depth=$DEPTH, timeout=${TIMEOUT}s, ${MAX_POSITIONS} per file, ${#EPD_FILES[@]} files"

# Add new result to the file
echo "| $TIMESTAMP | $GIT_COMMIT | $EPD_DESCRIPTION | $POSITIONS_TESTED | $SCORE | $MAX_SCORE | ${SCORE_PERCENT}% | $STS_RATING | $CORRECT_MOVES | $DEPTH | ${TIMEOUT}s | $AVG_TIME | ${TOTAL_TIME}s | $NPS_RAW | $AVG_DEPTH | $NOTES |" >> "$RESULTS_FILE"

# Display aggregated summary
echo "Aggregated STS Results Summary"
echo "=============================="
echo "Files processed: ${#EPD_FILES[@]} ($(basename "$EPD_DIR")/STS*.epd)"
echo "Total positions: $POSITIONS_TESTED"
echo "Total score: $SCORE/$MAX_SCORE (${SCORE_PERCENT}%)"
echo "Correct moves: $CORRECT_MOVES/$POSITIONS_TESTED"
echo "STS Rating: $STS_RATING"
echo "Total time: ${TOTAL_TIME}s"
echo "Average time per position: $AVG_TIME"
echo "Average depth: $AVG_DEPTH"
echo "Nodes per second: $NPS_RAW"
echo ""
echo "Results saved to $RESULTS_FILE"