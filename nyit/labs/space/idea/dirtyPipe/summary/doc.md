It looks like the `dirtypipezoo/cve-2022-0847` image is no longer available on Docker Hub, which is why you're seeing the "pull access denied" error.

---

## ✅ Ready-Made Alternatives

### 1. **Use a Vulnerable VM Instead**

Setting up a VM is straightforward and gives you full control over the unpatched kernel:

1. Download a distro with a vulnerable kernel (e.g., **Ubuntu 21.10** with kernel \~5.13).
2. Install it in VirtualBox or QEMU.
3. Disable automatic updates to prevent patching.
4. Compile or download the Dirty Pipe **PoC exploit** (e.g., from Max Kellermann).
5. Run the PoC — you'll get successful root escalation or file overwrite ([jfrog.com][1], [securitylabs.datadoghq.com][2]).

This is the most reliable method since you're guaranteed to be on the right kernel.

---

### 2. **Custom Docker + QEMU Image**

You can build your own test image similar to what `dirtypipezoo` provided:

* Create a **Dockerfile** that sets up a small QEMU VM image using a vulnerable kernel (5.8–5.15.24).
* Install the exploit PoC inside the VM.
* Run the container with `--privileged` and appropriate QEMU parameters to boot into the unpatched environment.

This method gives you flexibility—but it's more work than spinning up a VM.

---

### 3. **Community-Maintained Vulnerability Images**

While the original `dirtypipezoo` image is gone, it's worth checking other community sources or repos such as:

* GitHub searches for `dirty pipe docker image`
* Other security labs or CTF-focused images

⚠️ Always verify kernel version inside these images (must be unpatched: 5.8–5.15.24) and run locally with `--privileged`.

---

## 🧭 Summary

* ✅ No official Docker Hub image currently available.
* 🔧 **Best option:** Create a small vulnerable **VM** with kernel ≤ 5.15.24 and test freely.
* 🛠️ Advanced users: build your own Docker+QEMU image with a vulnerable kernel and PoC preloaded.

---

Would you like help setting up a **step-by-step lab in VirtualBox/QEMU**, or instructions to create a custom Docker+QEMU image?

[1]: https://jfrog.com/blog/dirtypipe-cve-2022-0847-the-new-dirtycow/?utm_source=chatgpt.com "DirtyPipe (CVE-2022-0847) - the new DirtyCoW? - JFrog"
[2]: https://securitylabs.datadoghq.com/articles/dirty-pipe-vulnerability-overview-and-remediation/?utm_source=chatgpt.com "The Dirty Pipe vulnerability: Overview, detection, and remediation"
