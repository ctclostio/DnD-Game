#Requires -Version 5.1
<#
.SYNOPSIS
    Finds and converts non-UTF-8 files in a Git repository to UTF-8.
.DESCRIPTION
    This script identifies all text files in the current git repository that are
    not valid UTF-8. It assumes the original encoding is Windows-1252 (Code Page 1252)
    and converts the files to UTF-8 (without BOM), overwriting the originals.
    A summary of converted files is printed upon completion.
.NOTES
    This script must be run from the root of a git repository.
    It requires Git to be installed and accessible in the system's PATH.
#>
[CmdletBinding()]
param()

begin {
    Write-Host "Starting UTF-8 encoding audit..."
    Write-Host "---------------------------------"

    # Define common binary file extensions to exclude from processing.
    $excludedExtensions = @('.png', '.jpg', '.jpeg', '.gif', '.ico', '.woff', '.woff2', '.eot', '.ttf', '.otf', '.zip', '.gz', '.tar', '.pdf', '.exe', '.dll', '.so', '.a', '.o', '.bin', '.dmg')

    # Get the Windows-1252 encoding.
    $windows1252 = [System.Text.Encoding]::GetEncoding(1252)

    # Get the UTF-8 encoding (without BOM).
    # Using this constructor ensures no Byte Order Mark (BOM) is written.
    $utf8NoBom = [System.Text.UTF8Encoding]::new($false)

    $convertedFiles = [System.Collections.Generic.List[string]]::new()
    $errorFiles = [System.Collections.Generic.List[string]]::new()
}

process {
    # Helper function to check if Git considers a file to be binary.
    function IsGitBinary($filePath) {
        try {
            $attributes = git check-attr text -- $filePath
            if ($attributes -like "*: text: unset") {
                return $true
            }
            return $false
        }
        catch {
            # If git command fails, fall back to false.
            Write-Warning "Could not run 'git check-attr' on '$($filePath)'. Assuming text file."
            return $false
        }
    }

    # Use 'git ls-files' as the most reliable way to get all tracked files.
    $files = git ls-files | ForEach-Object { Get-Item -Path $_ }

    foreach ($file in $files) {
        # Skip directories, excluded file types, and files git considers binary
        if ($file.PSIsContainer -or $excludedExtensions -contains $file.Extension) {
            continue
        }

        if (IsGitBinary($file.FullName)) {
            Write-Verbose "Skipping binary file identified by git: $($file.FullName)"
            continue
        }

        try {
            # Attempt to read the file as strict UTF-8.
            # If it contains invalid byte sequences, this will throw an exception.
            $null = Get-Content -Path $file.FullName -Encoding UTF8 -ErrorAction Stop
        }
        catch [System.Text.DecoderFallbackException] {
            # This exception block is entered if the file is NOT valid UTF-8.
            Write-Host "Converting non-UTF-8 file: $($file.FullName)"

            try {
                # Read the file assuming Windows-1252 encoding.
                $content = Get-Content -Path $file.FullName -Raw -Encoding $windows1252 -ErrorAction Stop

                # Write the content back as UTF-8 (without BOM).
                # Set-Content with -NoNewline is used to preserve ending newlines correctly when using -Raw.
                Set-Content -Path $file.FullName -Value $content -Encoding $utf8NoBom -NoNewline -Force
                
                $convertedFiles.Add($file.FullName)
            }
            catch {
                Write-Warning " -> ERROR: Failed to convert $($file.FullName). Manual inspection required."
                $errorFiles.Add($file.FullName)
            }
        }
        catch {
            # Catch other potential reading errors.
            Write-Warning " -> ERROR: Could not read file $($file.FullName). Skipping."
            $errorFiles.Add($file.FullName)
        }
    }
}

end {
    Write-Host "---------------------------------"
    if ($convertedFiles.Count -eq 0) {
        Write-Host "Audit complete. No files were converted. All text files appear to be valid UTF-8."
    }
    else {
        Write-Host "Audit complete. The following files were converted to UTF-8:" -ForegroundColor Green
        foreach ($file in $convertedFiles) {
            Write-Host "  - $file"
        }
        Write-Host "Please review the changes with 'git diff' before committing."
    }

    if ($errorFiles.Count -gt 0) {
        Write-Host "`nThe following files encountered errors and were NOT converted:" -ForegroundColor Yellow
        foreach ($file in $errorFiles) {
            Write-Host "  - $file"
        }
    }
}
