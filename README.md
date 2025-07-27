[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Jenvy - Developer Kit Manager

<div style="display: flex; align-items: center">
<img src="assets/icons/jenvy_white.svg" alt="logo" height="400" />
<strong>A professional solution for centralized OpenJDK distributions management
</strong>
</div>

---

Jenvy is a command-line application designed to simplify the installation, management, and switching between different versions of OpenJDK on Windows systems. The tool supports major public providers (Adoptium, Azul Zulu, BellSoft Liberica) and private enterprise repositories.

> **⚠️ Important:** This is a personal and independent open source project. I am not affiliated with Oracle Corporation or its products. Jenvy is a management tool for third-party OpenJDK distributions and does not include, distribute, or modify any Oracle software.

---

## Main Features

### Multi-Provider Management

-   **Public Providers**: Native integration with Adoptium (Eclipse Temurin), Azul Zulu, and BellSoft Liberica
-   **Private Repositories**: Complete support for custom enterprise JDK distributions
-   **Flexible Configuration**: Management through local configuration files or environment variables

### Core Operations

-   **Remote Exploration**: Search and display available JDK versions with advanced filters
-   **Intelligent Download**: Automatic download with system architecture detection
-   **Automatic Extraction**: Option for immediate extraction upon download completion
-   **Local Management**: Display and administration of installed JDK versions
-   **Automatic Switching**: Active JDK version switching with automatic privilege elevation

### Advanced Features

-   **Auto-completion**: Native support for Bash, PowerShell, and Command Prompt
-   **Smart Filters**: Automatic selection based on LTS criteria, major versions, and latest patches
-   **PATH Management**: Integrated tools for system variables repair and maintenance
-   **Safe Removal**: Controlled deletion with security confirmations for destructive operations

---

## Installation

### Windows Distribution

1. Download the `jenvy-installer.exe` file from the releases section
2. Run the installer with administrator privileges
3. The `jenvy` command will be available globally in all terminals

### Compilation from Source

```bash
# Clone the repository
git clone https://github.com/MarcoAntonioRussoDEV/Jenvy.git
cd Jenvy

# Complete build with installer (requires Inno Setup)
./build.bat
```

---

## Usage Guide

### Exploring Available Versions

```bash
# Display versions from default provider (Adoptium)
jenvy remote-list

# Explore specific providers
jenvy remote-list --provider=azul
jenvy remote-list --provider=liberica
jenvy remote-list --provider=private

# Advanced filters
jenvy remote-list --lts-only          # Only Long Term Support versions
jenvy remote-list --major-only        # Only major versions
jenvy remote-list --latest            # Only the latest versions
jenvy remote-list --all               # All versions from all providers
```

### Download and Installation

```bash
# Download a specific version
jenvy download 21

# The system will automatically ask if you want to extract the archive:
# [?] Do you want to extract the archive now? (Y/n):
# - Y/y/Enter: Immediate automatic extraction
# - n/N: Download only, manual extraction later

# Manual extraction of already downloaded archives
jenvy extract JDK-21.0.1+12
```

### Managing Installed Versions

```bash
# Display installed versions
jenvy list

# Activate a specific version (requires admin privileges)
jenvy use 21


### Private Repository Administration
```

### Private Repositories

```bash
# Configure private repository
jenvy configure-private https://repository.company.com/jdk YOUR_TOKEN

# Display current configuration
jenvy config-show

# Reset configuration
jenvy config-reset
```

### Removal and Maintenance

```bash
# Remove specific version
jenvy remove 17

# Complete removal (with security confirmation)
jenvy remove --all

# Repair system variables
jenvy fix-path
```

---

## Advanced Configuration

### Private Repositories

The system supports two configuration modes for private repositories:

#### Configuration File

Path: `%USERPROFILE%\.jenvy\config.json`

```json
{
    "private": {
        "endpoint": "https://repository.company.com/api/jdk",
        "token": "your-auth-token"
    }
}
```

#### Environment Variables

```bash
set JENVY_PRIVATE_ENDPOINT=https://repository.company.com/api/jdk
set JENVY_PRIVATE_TOKEN=your-auth-token
```

### Private Repository API Structure

The system requires private repositories to expose a REST endpoint that returns a JSON array with available JDK versions. The endpoint can support authentication via `Authorization: Bearer <token>` header.

#### Endpoint Specification

**URL:** `GET {endpoint}/api/jdk` or configured endpoint  
**Headers:** `Authorization: Bearer {token}`  
**Content-Type:** `application/json`

#### JSON Response Format

```json
[
    {
        "version": "11.0.21",
        "download": "https://repository.company.com/private-jdk/openjdk-11.0.21.zip",
        "os": "windows",
        "arch": "x64",
        "lts": true
    },
    {
        "version": "17.0.15",
        "download": "https://repository.company.com/private-jdk/openjdk-17.0.15.zip",
        "os": "windows",
        "arch": "x64",
        "lts": true
    },
    {
        "version": "21.0.7",
        "download": "https://repository.company.com/private-jdk/openjdk-21.0.7.zip",
        "os": "windows",
        "arch": "x64",
        "lts": true
    },
    {
        "version": "22.0.2",
        "download": "https://repository.company.com/private-jdk/openjdk-22.0.2.zip",
        "os": "windows",
        "arch": "x64",
        "lts": false
    }
]
```

#### Required Fields

| Field      | Type    | Description                                   | Accepted Values                                          |
| ---------- | ------- | --------------------------------------------- | -------------------------------------------------------- |
| `version`  | String  | JDK semantic version                          | Format: `major.minor.patch` or `major.minor.patch+build` |
| `download` | String  | Direct URL for JDK archive download           | Valid HTTPS URL                                          |
| `arch`     | String  | CPU Architecture                              | `x64`, `x32`, `aarch64`                                  |
| `lts`      | Boolean | Indicates if it's a Long Term Support version | `true`, `false`                                          |

#### Server Implementation Example

```javascript
// Node.js/Express endpoint example
app.get("/api/jdk", authenticateToken, (req, res) => {
    const jdkVersions = [
        {
            version: "11.0.21",
            download:
                "https://repository.company.com/private-jdk/openjdk-11.0.21.zip",
            arch: "x64",
            lts: true,
        },
        // ... other versions
    ];

    res.json(jdkVersions);
});

function authenticateToken(req, res, next) {
    const authHeader = req.headers["authorization"];
    const token = authHeader && authHeader.split(" ")[1];

    if (!token || !isValidToken(token)) {
        return res.sendStatus(401);
    }

    next();
}
```

## Windows Privilege Management

### Automatic UAC Elevation

The `jenvy use` command automatically requests privilege elevation through Windows UAC dialog to:

-   Modify the `JAVA_HOME` system variable
-   Update the `PATH` system variable
-   Ensure persistence of changes for all users

**Operational flow:**

1. Execute command `jenvy use <version>`
2. Automatic privilege elevation request
3. User confirmation via UAC dialog
4. Apply changes with administrative privileges

---

## 💖 Support the Project

Jenvy is an open source project developed in my spare time. If you find this tool useful and want to support its development, consider a donation:

### 🎯 Donation Options

-   **GitHub Sponsors**: [Sponsor on GitHub](https://github.com/sponsors/MarcoAntonioRussoDEV)
-   **Ko-fi**: [Support on Ko-fi](https://ko-fi.com/marcoantoniorussodev)
-   **PayPal**: [Donate via PayPal](https://paypal.me/Ocrama94)

### 🚀 How donations are used

Donations help to:

-   Keep the project active and updated
-   Add new features requested by the community
-   Improve documentation and tests

### 🤝 Other ways to contribute

Even if you can't donate, you can support the project:

-   ⭐ Star the repository on GitHub
-   🐛 Report bugs and issues
-   💡 Suggest new features
-   📖 Improve documentation
-   🔧 Contribute with pull requests

---

## 📄 License

This project is released under **MIT License**.

```
MIT License

Copyright (c) 2025 Marco Antonio Russo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### 🔒 Disclaimer and Liability

-   **No affiliation**: This project is not affiliated, approved, or sponsored by Oracle Corporation
-   **Third-party software**: Jenvy manages OpenJDK distributions provided by third-party providers (Eclipse Adoptium, Azul, BellSoft)
-   **Use at your own risk**: The software is provided "as-is" without warranties of any kind
-   **User responsibility**: The user is responsible for compliance with downloaded JDK licenses
-   **Trademarks**: Java and OpenJDK are registered trademarks of Oracle Corporation

---

## 🤝 Contributing

Contributions, bug reports, and feature requests are welcome!

### 📋 How to contribute

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### 🐛 Report Bugs

Open an [issue on GitHub](https://github.com/MarcoAntonioRussoDEV/Jenvy/issues) including:

-   Windows version used
-   Jenvy version (`jenvy --version`)
-   Detailed problem description
-   Error log (if available)
-   Steps to reproduce the bug

### 💡 Request Features

For new features, open a [discussion on GitHub](https://github.com/MarcoAntonioRussoDEV/Jenvy/discussions) specifying:

-   Specific use case
-   Desired behavior
-   Any alternatives considered

---

## 📞 Contacts

-   **GitHub**: [@MarcoAntonioRussoDEV](https://github.com/MarcoAntonioRussoDEV)
-   **Email**: marcoantoniorusso94@gmail.com

---

![signature](assets/images/SVG_GRADIENT_WHITE.svg)

---
