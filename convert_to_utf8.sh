#!/bin/bash
#
# This script finds all non-UTF-8 encoded text files in the current repository,
# assumes they are WINDOWS-1252, and converts them to UTF-8 in-place.
# It prints a list of the files that were converted.
#
# It requires the 'iconv' utility, which is standard on Linux and macOS.
# For Windows, you can run this script via Git Bash or WSL.

set -e

# We don't want to check files inside .git directory.
# We also want to skip binary files. The 'file' command is used for this,
# and we exclude common binary extensions for performance and safety.
EXCLUDED_EXTENSIONS='(png|jpg|jpeg|gif|ico|woff|woff2|eot|ttf|otf|zip|gz|tar|pdf|exe|dll|so|a|o|bin|dmg)'

echo "Starting UTF-8 encoding audit..."
echo "---------------------------------"

converted_files=()

# Use 'git ls-files' to get a list of all tracked files in the repo.
# This is safer and more accurate than 'find'.
# IFS is set to handle filenames with spaces correctly.
IFS=$'\n'
for f in $(git ls-files); do
  # Skip excluded extensions
  if [[ "$f" =~ \.${EXCLUDED_EXTENSIONS}$ ]]; then
    continue
  fi

  # Skip directories
  if [ -d "$f" ]; then
    continue
  fi

  # Check if the file is text-based using 'file' command
  # The -b option removes the filename from the output.
  if ! file -b --mime-type "$f" | grep -q "text/"; then
    continue
  fi

  # Try to read the file as UTF-8. 'iconv' will exit with an error
  # if it encounters invalid UTF-8 sequences.
  if ! iconv -f UTF-8 -t UTF-8 -o /dev/null "$f" >/dev/null 2>&1; then
    echo "Converting non-UTF-8 file: $f"

    # Assume WINDOWS-1252 and convert to UTF-8.
    # The -c flag discards characters that cannot be converted.
    # Create a temporary file for the conversion.
    temp_file=$(mktemp)
    if iconv -f WINDOWS-1252 -t UTF-8 -c "$f" > "$temp_file"; then
      # Overwrite the original file with the converted content
      mv "$temp_file" "$f"
      converted_files+=("$f")
    else
      echo "  -> ERROR: Failed to convert $f. Manual inspection required."
      rm -f "$temp_file"
    fi
  fi
done

echo "---------------------------------"
if [ ${#converted_files[@]} -eq 0 ]; then
  echo "Audit complete. No files were converted. All text files appear to be valid UTF-8."
else
  echo "Audit complete. The following files were converted to UTF-8:"
  for file in "${converted_files[@]}"; do
    echo "  - $file"
  done
  echo "Please review the changes with 'git diff' before committing."
fi
