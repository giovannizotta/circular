#!/bin/bash
set -x
# Get the directory of the script
script_dir=$(dirname -- "$(readlink -f -- "$0")")

# get latest version
version=$(curl -s https://api.github.com/repos/giovannizotta/circular/releases/latest | grep -o '"tag_name": *"[^"]*"' | grep -o '"[^"]*"$' | tr -d '"')

get_platform_file_end() {
    machine=$(uname -m)
    kernel=$(uname -s)

    case $kernel in
        Darwin)
	    case $machine in
		x86_64)
	            echo 'darwin-amd64.tar.gz'
		    ;;
		arm64)
		    echo 'darwin-arm64.tar.gz'
		    ;;
		*)
		    echo "Unsupported release-architecture: $machine" >&2
		    exit 1
		    ;;
	    esac
            ;;
        Linux)
            case $machine in
                x86_64)
                    echo 'linux-amd64.tar.gz'
                    ;;
                aarch64)
                    echo 'linux-arm64.tar.gz'
                    ;;
                *)
                    echo "Unsupported release-architecture: $machine" >&2
                    exit 1
                    ;;
            esac
            ;;
        *)
            echo "Unsupported OS: $kernel" >&2
            exit 1
            ;;
    esac
}
platform_file_end=$(get_platform_file_end)
archive_file=circular-$version-$platform_file_end

github_url="https://github.com/giovannizotta/circular/releases/download/$version/$archive_file"


# Download the archive using curl
if ! curl -L "$github_url" -o "$script_dir/$archive_file"; then
    echo "Error downloading the file from $github_url" >&2
    exit 1
fi

# Extract the contents
if [[ $archive_file == *.tar.gz ]]; then
    if ! tar -xzvf "$script_dir/$archive_file" -C "$script_dir"; then
        echo "Error extracting the contents of $archive_file" >&2
        exit 1
    fi
elif [[ $archive_file == *.zip ]]; then
    if ! unzip "$script_dir/$archive_file" -d "$script_dir"; then
        echo "Error extracting the contents of $archive_file" >&2
        exit 1
    fi
else
    echo "Unknown archive format or unsupported file extension: $archive_file" >&2
    exit 1
fi

