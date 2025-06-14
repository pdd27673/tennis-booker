#!/bin/bash

# MongoDB Index Optimization Script
# This script automates the process of analyzing and optimizing MongoDB indexes

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ANALYSIS_FILE="mongodb_index_analysis.json"
RESULTS_FILE="mongodb_optimization_results.json"
DRY_RUN=false
VALIDATE_ONLY=false

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
MongoDB Index Optimization Script

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -a, --analysis-file     Specify analysis output file (default: mongodb_index_analysis.json)
    -r, --results-file      Specify optimization results file (default: mongodb_optimization_results.json)
    -d, --dry-run           Only run analysis, don't apply optimizations
    -v, --validate          Only run validation tests
    --analyze-only          Only run database analysis
    --optimize-only         Only run optimization (requires existing analysis file)

ENVIRONMENT VARIABLES:
    MONGO_URI              MongoDB connection string (default: mongodb://admin:YOUR_PASSWORD@localhost:27017)
    MONGO_DB_NAME          Database name (default: tennis_booking)

EXAMPLES:
    # Full optimization process
    $0

    # Only analyze database
    $0 --analyze-only

    # Only apply optimizations using existing analysis
    $0 --optimize-only

    # Dry run (analysis only)
    $0 --dry-run

    # Validate tools without database connection
    $0 --validate

    # Custom file names
    $0 -a custom_analysis.json -r custom_results.json

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -a|--analysis-file)
            ANALYSIS_FILE="$2"
            shift 2
            ;;
        -r|--results-file)
            RESULTS_FILE="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -v|--validate)
            VALIDATE_ONLY=true
            shift
            ;;
        --analyze-only)
            ANALYZE_ONLY=true
            shift
            ;;
        --optimize-only)
            OPTIMIZE_ONLY=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Function to check if tools are compiled
check_tools() {
    print_status "Checking if tools are compiled..."
    
    local tools_missing=false
    
    if [[ ! -f "analyze_indexes" ]]; then
        print_warning "analyze_indexes not found, compiling..."
        if ! go build -o analyze_indexes analyze_indexes.go; then
            print_error "Failed to compile analyze_indexes"
            exit 1
        fi
        print_success "analyze_indexes compiled successfully"
    fi
    
    if [[ ! -f "optimize_indexes" ]]; then
        print_warning "optimize_indexes not found, compiling..."
        if ! go build -o optimize_indexes optimize_indexes.go; then
            print_error "Failed to compile optimize_indexes"
            exit 1
        fi
        print_success "optimize_indexes compiled successfully"
    fi
    
    if [[ ! -f "validate_tools" ]]; then
        print_warning "validate_tools not found, compiling..."
        if ! go build -o validate_tools validate_tools.go; then
            print_error "Failed to compile validate_tools"
            exit 1
        fi
        print_success "validate_tools compiled successfully"
    fi
}

# Function to validate tools
validate_tools() {
    print_status "Running validation tests..."
    if ./validate_tools; then
        print_success "All validation tests passed"
        return 0
    else
        print_error "Validation tests failed"
        return 1
    fi
}

# Function to run database analysis
run_analysis() {
    print_status "Running MongoDB database analysis..."
    print_status "Output file: $ANALYSIS_FILE"
    
    if ./analyze_indexes "$ANALYSIS_FILE"; then
        print_success "Database analysis completed successfully"
        print_status "Analysis results saved to: $ANALYSIS_FILE"
        return 0
    else
        print_error "Database analysis failed"
        return 1
    fi
}

# Function to apply optimizations
apply_optimizations() {
    print_status "Applying MongoDB index optimizations..."
    print_status "Using analysis file: $ANALYSIS_FILE"
    print_status "Results file: $RESULTS_FILE"
    
    if [[ ! -f "$ANALYSIS_FILE" ]]; then
        print_error "Analysis file not found: $ANALYSIS_FILE"
        print_error "Run analysis first or specify correct file with -a option"
        return 1
    fi
    
    if ./optimize_indexes "$ANALYSIS_FILE"; then
        print_success "Index optimization completed successfully"
        print_status "Optimization results saved to: $RESULTS_FILE"
        return 0
    else
        print_error "Index optimization failed"
        return 1
    fi
}

# Function to show environment info
show_env_info() {
    print_status "Environment Configuration:"
    echo "  MONGO_URI: ${MONGO_URI:-mongodb://admin:YOUR_PASSWORD@localhost:27017}"
    echo "  MONGO_DB_NAME: ${MONGO_DB_NAME:-tennis_booking}"
    echo "  Analysis file: $ANALYSIS_FILE"
    echo "  Results file: $RESULTS_FILE"
    echo ""
}

# Main execution
main() {
    echo "🚀 MongoDB Index Optimization Tool"
    echo "=================================="
    echo ""
    
    # Check if we're in the right directory
    if [[ ! -f "analyze_indexes.go" ]]; then
        print_error "analyze_indexes.go not found. Please run this script from the mongodb scripts directory."
        exit 1
    fi
    
    # Show environment info
    show_env_info
    
    # Validate only mode
    if [[ "$VALIDATE_ONLY" == true ]]; then
        check_tools
        validate_tools
        exit $?
    fi
    
    # Check and compile tools
    check_tools
    
    # Run validation first
    if ! validate_tools; then
        print_error "Validation failed, aborting"
        exit 1
    fi
    
    # Analyze only mode
    if [[ "$ANALYZE_ONLY" == true ]]; then
        run_analysis
        exit $?
    fi
    
    # Optimize only mode
    if [[ "$OPTIMIZE_ONLY" == true ]]; then
        apply_optimizations
        exit $?
    fi
    
    # Dry run mode (analysis only)
    if [[ "$DRY_RUN" == true ]]; then
        print_warning "Dry run mode: Only running analysis"
        run_analysis
        exit $?
    fi
    
    # Full optimization process
    print_status "Running full optimization process..."
    
    # Step 1: Analyze database
    if ! run_analysis; then
        exit 1
    fi
    
    echo ""
    print_status "Analysis complete. Review the recommendations in $ANALYSIS_FILE"
    
    # Ask for confirmation before applying optimizations
    echo ""
    read -p "Do you want to apply the optimizations? (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Step 2: Apply optimizations
        if apply_optimizations; then
            echo ""
            print_success "🎉 MongoDB optimization completed successfully!"
            print_status "Check $RESULTS_FILE for detailed results"
            print_status "Verify indexes in MongoDB with: db.collection.getIndexes()"
        else
            exit 1
        fi
    else
        print_status "Optimization cancelled by user"
        print_status "You can apply optimizations later with: $0 --optimize-only"
    fi
}

# Run main function
main "$@" 