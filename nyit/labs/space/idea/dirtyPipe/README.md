# What is this

https://hub.docker.com/r/iridium191/cve-2022-0847

https://sysdig.com/blog/cve-2022-0847-dirty-pipe-sysdig/
https://www.aquasec.com/blog/cve-2022-0847-dirty-pipe-linux-vulnerability/

### 🛡️ CVE-2022-0847 — "Dirty Pipe" Vulnerability

**CVE-2022-0847**, nicknamed **"Dirty Pipe"**, is a **critical privilege escalation vulnerability** in the **Linux kernel**, discovered by security researcher Max Kellermann in early 2022.

---

## 🧨 What Is It?

It allows **unprivileged local users** to **overwrite read-only files**, including system binaries, **by exploiting a bug in the Linux pipe implementation**. This can lead to:

* Full **root privilege escalation**
* Tampering with **sensitive system files** like `/etc/passwd`
* **Container breakout** in shared-kernel environments

---

## 🧬 Technical Summary

* A **pipe buffer** implementation flaw allows users to **inject arbitrary data** into page cache-backed read-only files.
* This violates the **read-only nature of file descriptors**, meaning files you normally can’t write to can be **silently modified in memory**.
* The kernel fails to properly handle flags (`PIPE_BUF_FLAG_CAN_MERGE`) during the use of `splice()` and `write()` on pipes.

---

## 📅 Affected Systems

| Component        | Affected Versions                                              |
| ---------------- | -------------------------------------------------------------- |
| **Linux Kernel** | Versions **5.8 to 5.16.10**, 5.17-rc1 and rc2                  |
| **Platforms**    | Most modern Linux distros (Ubuntu, Debian, Fedora, Arch, etc.) |

This includes:

* Linux desktops
* Servers
* Containers (Docker, LXC)
* Cloud environments

---

## 🧪 Example Exploits

Researchers and PoC writers have demonstrated:

* Modifying `/etc/passwd` to add a root shell
* Replacing binaries like `/bin/su`
* Overwriting SSH keys or configuration files

> 🔐 **This is not just a denial-of-service bug — it gives full root access!**

---

## 🛠️ Mitigation & Fix

### ✅ Patched Kernels:

* Fixed in **5.16.11**, **5.15.25**, **5.10.102**, and newer LTS versions.
* Most distros issued **emergency patches** in early March 2022.

### 🔒 Recommendations:

1. **Update your kernel immediately** to a version with the fix.
2. Use `uname -r` to check your kernel version.
3. Reboot into the updated kernel to apply changes.
4. If unable to patch, consider using tools like:

   * `AppArmor` / `SELinux` to restrict access
   * Unprivileged user namespaces mitigation (`sysctl -w kernel.unprivileged_userns_clone=0`)

---

## ❗ Impact Summary

| Attribute           | Value                     |
| ------------------- | ------------------------- |
| **CVE**             | CVE-2022-0847             |
| **Nickname**        | Dirty Pipe                |
| **Type**            | Privilege Escalation      |
| **Access Required** | Local                     |
| **Exploitability**  | Easy with public PoCs     |
| **Severity**        | High (CVSS \~7.8–8.8)     |
| **Similar to**      | Dirty COW (CVE-2016-5195) |

---

## 📚 References

* Original report: [https://dirtypipe.cm4all.com/](https://dirtypipe.cm4all.com/)
* GitHub PoC: [https://github.com/Arinerron/CVE-2022-0847-DirtyPipe-Exploit](https://github.com/Arinerron/CVE-2022-0847-DirtyPipe-Exploit)

---

Let me know if you want a hands-on demo or mitigation script.


Yes, there are Docker images and GitHub repositories available to **set up an environment and reproduce CVE-2022-0847 ("Dirty Pipe")**, but with some important **caveats and warnings**.

---

## ⚠️ Before You Proceed

**This is a critical, real-world kernel exploit**. Please ensure:

1. 🧪 **You're testing in a safe, isolated environment** (e.g., a VM or container **not connected to production**).
2. 🔐 **Never run this on your host or publicly exposed systems**.
3. ⚖️ **Understand the legal/ethical risks** — use only for educational or research purposes.

---

## 🐳 Docker Image-Based CVE Reproduction

### ✅ Option 1: Use Community Repos with Dockerfile

You can find public repos with Dockerfiles for vulnerable Linux kernels and Dirty Pipe PoC:

### Example Repo:

**[https://github.com/hakivvi/CVE-2022-0847-dirtypipe-docker](https://github.com/hakivvi/CVE-2022-0847-dirtypipe-docker)**

> 🔗 GitHub: [hakivvi/CVE-2022-0847-dirtypipe-docker](https://github.com/hakivvi/CVE-2022-0847-dirtypipe-docker)

### ✅ What it does:

* Runs a **vulnerable kernel version** (via container or VM)
* Includes **C code PoC** that demonstrates the exploit
* Uses a base image like Debian or Ubuntu

---

## 🧱 Manual Setup (Safer + More Control)

If you'd rather build your own:

### 1. **Base Dockerfile (vulnerable kernel)**

```Dockerfile
FROM ubuntu:21.10

RUN apt update && apt install -y build-essential wget curl

# Add vulnerable kernel headers or simulate dirty pipe if running in container
COPY exploit.c /exploit.c

CMD ["/bin/bash"]
```

> Note: Docker containers share the **host kernel**, so the kernel must be vulnerable **on your host** or in a nested VM.

---

## 🚫 Limitation: Docker shares host kernel

> ❗ Dirty Pipe exploits **Linux kernel**, and containers **cannot downgrade or run a separate kernel**.

### ✅ Solution: Use a VM or Nested Virtualization

You can:

1. Use Docker to build PoC tooling and run it inside...
2. ...a **VM with a vulnerable kernel**, e.g., QEMU with Linux 5.10–5.16.10

---

## 🛠️ Steps to Reproduce Dirty Pipe

1. Spin up a **VM or Docker container** with kernel 5.10–5.16.10.
2. Compile PoC exploit:

```bash
gcc exploit.c -o exploit
```

3. Run with a test file to overwrite (e.g., `/etc/passwd` in test-only sandbox).
4. Observe overwritten read-only file content.

---

## 🧪 PoC Repos

| Repository                                                                                                | Description                         |
| --------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| [Arinerron/CVE-2022-0847-DirtyPipe-Exploit](https://github.com/Arinerron/CVE-2022-0847-DirtyPipe-Exploit) | Well-documented exploit             |
| [makelinux/CVE-2022-0847](https://github.com/makelinux/CVE-2022-0847)                                     | Includes Docker + QEMU instructions |
| [hakivvi/CVE-2022-0847-dirtypipe-docker](https://github.com/hakivvi/CVE-2022-0847-dirtypipe-docker)       | Docker-based setup                  |
| [dirtypipezoo](https://github.com/knavesec/dirtypipezoo)                                                  | Collection of Dirty Pipe variants   |

---

## ✅ Safer Alternatives

If you can't test in Docker due to shared host kernel, use:

* ✅ **QEMU** to emulate a VM with an old kernel
* ✅ **Vagrant** with specific Linux boxes
* ✅ **Nested KVM/QEMU** inside your Linux VM

---

## 📌 Summary

| Task                 | Tool                                    |
| -------------------- | --------------------------------------- |
| Safe testing         | Use QEMU or a sandboxed VM              |
| Build environment    | Docker image with `build-essential`     |
| Host kernel required | Linux 5.8–5.16.10                       |
| Real-world PoC       | GitHub repos (e.g., hakivvi, Arinerron) |

---

Let me know if you want a working Dockerfile + PoC setup or help building a QEMU-based vulnerable testbed.


This is Max Kellermann's proof of concept for Dirty Pipe, but modified to overwrite root's password field in /etc/passwd and restore after popping a root shell.

**Side Note:** I do not claim any credit for finding this vulnerability or writing the proof of concept. This exploit is merely a small modification of Kellermann's proof of concept to enable quick/easy exploitation. Please read the original article on this extremely interesting vulnerability @ https://dirtypipe.cm4all.com/ when you get the opportunity. It really does deserve your time to understand it.

# How to use this

1. Compile with `./compile.sh` (assumes `gcc` is installed)
2. Run `./exploit` and it'll pop a root shell

# su: must be run from a terminal

If you get this error message:
1. Login as `root` with the password `aaron`.
2. Then, restore `/etc/passwd` by running `mv /tmp/passwd.bak /etc/passwd`

(oops sorry my laptop battery is dying and my charger broke so I don't have time to fix this the right now, sorry)

