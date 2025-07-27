# Alternative Inno Setup Installation Methods for GitHub Actions

Se la GitHub Action `crazy-max/ghaction-setup-innosetup` dovesse avere problemi, ecco alcune alternative robuste:

## Metodo 1: Chocolatey (Raccomandato)

```yaml
- name: 🏗️ Setup Inno Setup via Chocolatey
  shell: powershell
  run: |
      Write-Host "📥 Installing Inno Setup via Chocolatey..."
      choco install innosetup --yes --no-progress

      # Verify installation
      $isccPath = Get-Command iscc -ErrorAction SilentlyContinue
      if ($isccPath) {
          Write-Host "✅ Inno Setup installed successfully at: $($isccPath.Source)"
          iscc | Select-Object -First 5
      } else {
          Write-Host "❌ Inno Setup installation failed"
          exit 1
      }
```

## Metodo 2: Download e installazione manuale migliorata

```yaml
- name: 🏗️ Setup Inno Setup (Manual - Robust)
  shell: powershell
  run: |
      $url = "https://files.jrsoftware.org/is/6/innosetup-6.4.3.exe"
      $installer = "innosetup.exe"

      Write-Host "📥 Downloading Inno Setup from: $url"
      Invoke-WebRequest -Uri $url -OutFile $installer -UseBasicParsing

      Write-Host "🔧 Installing Inno Setup silently..."
      Start-Process -FilePath $installer -ArgumentList "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/SP-", "/NOCANCEL", "/NOICONS" -Wait

      # Wait and verify installation
      $timeout = 60
      $elapsed = 0
      $interval = 5

      do {
          Start-Sleep $interval
          $elapsed += $interval
          $isccExists = Test-Path "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe"
          
          if ($isccExists) {
              Write-Host "✅ Inno Setup installed successfully!"
              break
          }
          
          Write-Host "   ... waiting for installation ($elapsed/$timeout seconds)"
      } while ($elapsed -lt $timeout)

      if (-not $isccExists) {
          Write-Host "❌ Inno Setup installation failed or timed out"
          Get-ChildItem "${env:ProgramFiles(x86)}" | Where-Object { $_.Name -like "*Inno*" }
          exit 1
      }

      # Add to PATH for current session
      $env:PATH += ";${env:ProgramFiles(x86)}\Inno Setup 6"

      Write-Host "🔍 Inno Setup installation verified:"
      Get-ChildItem "${env:ProgramFiles(x86)}\Inno Setup 6"
```

## Metodo 3: Cache per velocizzare successive build

```yaml
- name: 📦 Cache Inno Setup
  uses: actions/cache@v3
  id: cache-innosetup
  with:
      path: "C:\\Program Files (x86)\\Inno Setup 6"
      key: innosetup-6.4.3-${{ runner.os }}

- name: 🏗️ Setup Inno Setup (if not cached)
  if: steps.cache-innosetup.outputs.cache-hit != 'true'
  uses: crazy-max/ghaction-setup-innosetup@v3
  with:
      version: "6.4.3"
```

## Metodo 4: Portable version

```yaml
- name: 🏗️ Setup Inno Setup (Portable)
  shell: powershell
  run: |
      $url = "https://files.jrsoftware.org/is/6/innosetup-6.4.3.exe"
      $tempDir = "$env:TEMP\innosetup-portable"
      $installer = "$tempDir\innosetup.exe"

      New-Item -ItemType Directory -Path $tempDir -Force
      Invoke-WebRequest -Uri $url -OutFile $installer -UseBasicParsing

      # Extract installer contents (alternative approach)
      # Some users prefer to manually extract the installer files
      # This is more complex but gives more control

      Write-Host "Installing to portable directory..."
      Start-Process -FilePath $installer -ArgumentList "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/SP-", "/DIR=$tempDir\innosetup" -Wait

      # Add to PATH
      $env:PATH += ";$tempDir\innosetup"

      # Verify
      if (Test-Path "$tempDir\innosetup\ISCC.exe") {
          Write-Host "✅ Portable Inno Setup ready!"
      }
```

## Troubleshooting Tips

1. **Timeout Issues**: Aumenta il timeout a 120-180 secondi
2. **Path Issues**: Usa sempre il percorso completo per ISCC.exe
3. **Silent Installation**: Aggiungi `/NOCANCEL` ai parametri
4. **Verification**: Controlla sempre che ISCC.exe esista prima di usarlo
5. **Alternative**: Usa Chocolatey come backup se GitHub Action fallisce

## Script di Build dell'installer robusto

```yaml
- name: 📦 Build Installer (Robust)
  shell: powershell
  run: |
      # Create release directory
      New-Item -ItemType Directory -Path "release" -Force

      # Update setup.iss with current version
      Write-Host "🔧 Updating setup.iss with version: $env:VERSION"
      (Get-Content "build\installer\setup.iss") -replace '#define MyAppVersion ".*"', "#define MyAppVersion `"$env:VERSION`"" | Set-Content "build\installer\setup.iss"

      # Verify version was updated
      Select-String "MyAppVersion" "build\installer\setup.iss"

      # Find ISCC.exe
      $isccPaths = @(
          "iscc",  # If in PATH
          "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
          "${env:ProgramFiles}\Inno Setup 6\ISCC.exe",
          "$env:TEMP\innosetup-portable\innosetup\ISCC.exe"
      )

      $isccFound = $false
      foreach ($path in $isccPaths) {
          try {
              if ($path -eq "iscc") {
                  $null = Get-Command iscc -ErrorAction Stop
                  $isccCmd = "iscc"
              } else {
                  if (Test-Path $path) {
                      $isccCmd = $path
                  }
              }
              Write-Host "✅ Found ISCC at: $isccCmd"
              $isccFound = $true
              break
          } catch {
              continue
          }
      }

      if (-not $isccFound) {
          Write-Host "❌ ISCC.exe not found in any expected location"
          exit 1
      }

      # Compile installer
      Write-Host "🔨 Compiling installer..."
      Set-Location "build\installer"
      & $isccCmd "setup.iss"

      if ($LASTEXITCODE -ne 0) {
          Write-Host "❌ Installer compilation failed with exit code: $LASTEXITCODE"
          exit 1
      }

      # Verify installer was created
      Set-Location "..\\.."
      Write-Host "✅ Installer build completed!"
      Get-ChildItem "release" -Recurse
```
