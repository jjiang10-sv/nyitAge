To change the extension marketplace in Google Antigravity (or other VS Code forks) from OpenVSX to the official Microsoft Visual Studio Marketplace, you need to point the gallery configuration to Microsoft's servers.

**Note:** Accessing the Microsoft Marketplace from non-Microsoft products (like Antigravity/VSCodium) technically violates Microsoft's Terms of Service, though it is physically possible and commonly done.

### Method 1: The "Easy" Way (If supported in your version)

Recent versions of Antigravity often expose this directly in the settings UI to make the switch easier.

1. Open **Settings** (`Ctrl + ,` or `Cmd + ,`).
2. Search for **"Marketplace"** or navigate to **Antigravity Settings > Editor**.
3. Look for fields named **Marketplace Item URL** and **Marketplace Gallery URL**.
4. Replace the existing URLs with the official Microsoft ones:
* **Service/Gallery URL:** `https://marketplace.visualstudio.com/_apis/public/gallery`
* **Item URL:** `https://marketplace.visualstudio.com/items`


5. **Restart** the IDE.

---

### Method 2: The "Hard" Way (Editing `product.json`)

If the GUI settings are locked or unavailable, you must edit the internal configuration file `product.json`. This file is located inside the application's installation folder.

#### 1. Locate `product.json`

* **Windows:** `C:\Program Files\Antigravity\resources\app\product.json`
* **macOS:** Right-click the Antigravity app icon → **Show Package Contents** → `Contents/Resources/app/product.json`
* **Linux:** `/usr/share/antigravity/resources/app/product.json` (or similar depending on install path).

#### 2. Edit the file

Open `product.json` in a text editor (you may need Administrator/Sudo permissions). Find the `"extensionsGallery"` section and replace it with the following:

```json
"extensionsGallery": {
    "serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
    "itemUrl": "https://marketplace.visualstudio.com/items",
    "cacheUrl": "https://vscode.blob.core.windows.net/gallery/index",
    "controlUrl": ""
}

```

#### 3. Clear Cache & Restart

1. Save the file.
2. Delete the folder `~/.antigravity/extensions` (optional, but recommended to avoid conflicts between OpenVSX and MS versions).
3. Restart Antigravity.

### Why do this?

By default, forks like Antigravity use **OpenVSX**, an open-source alternative registry. While great, it lacks some proprietary extensions like the official **C/C++ (ms-vscode)**, **Pylance**, or **Live Share**, which are exclusive to the Microsoft store. Switching allows you to install these directly.