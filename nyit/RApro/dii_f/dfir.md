### Digital Forensics & Incident Response (DFIR) Flashcards  

---

#### **1. Digital Evidence**  
**Q**: What is digital evidence?  
**A**: Information stored/transmitted digitally that may be valuable to an investigation.  

---

#### **2. Data vs. Metadata**  
**Q**: What is the difference between data and metadata?  
**A**:  
- **Data**: Information in digital form (e.g., file content).  
- **Metadata**: Data about data (e.g., file creation time, author).  

---

#### **3. File Systems**  
**Q**: Name three common file systems and their typical uses.  
**A**:  
- **NTFS**: Windows systems.  
- **APFS**: Modern macOS.  
- **exFAT/FAT32**: USB devices.  

---

#### **4. Slack Space**  
**Q**: What are the two types of slack space in a cluster?  
**A**:  
- **RAM Slack**: Unused space between the end of a file and the end of the sector.  
- **File Slack**: Unused sectors in the last cluster occupied by a file.  

---

#### **5. Order of Volatility**  
**Q**: List the order of evidence collection to minimize data loss.  
**A**:  
1. CPU registers/cache (rarely feasible).  
2. RAM → 3. Router data → 4. Temp filesystems → 5. Hard drive → 6. Remote data → 7. Network config → 8. Offsite backups.  

---

#### **6. ACPO Principles**  
**Q**: What is Principle 2 of ACPO?  
**A**: If accessing original data is necessary, the person must be competent and explain the relevance/implications of their actions.  

---

#### **7. Locard’s Exchange Principle**  
**Q**: How does Locard’s Principle apply to digital devices?  
**A**: Any interaction leaves traces (e.g., opening a file creates access logs). Non-action or anti-forensic actions may also leave evidence (e.g., absence of expected data).  

---

#### **8. Hashing**  
**Q**: Why is hashing critical in forensics?  
**A**: It verifies data integrity (e.g., MD5/SHA256). A single change alters the hash. Hash collisions are rare but possible.  

---

#### **9. Deleting a File**  
**Q**: What happens when a file is deleted (not sent to Recycle Bin)?  
**A**:  
- File system metadata is marked reusable.  
- Clusters are marked unallocated.  
- Recovery is possible until overwritten.  

---

#### **10. Sources of Digital Evidence**  
**Q**: Name four sources beyond computers.  
**A**:  
- Mobile devices (Android/iOS).  
- Cloud services.  
- IoT devices (e.g., smart cameras).  
- Vehicles (GPS, paired device data).  

---

#### **11. DFIR Roles**  
**Q**: What is the primary responsibility of a SOC Analyst?  
**A**: Monitor events, detect incidents. Key skills: network defense, logging, endpoint monitoring.  

---

#### **12. Incident Response Phases**  
**Q**: What occurs during the *Containment* phase?  
**A**: Limit damage while balancing containment steps against attacker access (e.g., isolating infected systems).  

---

#### **13. Write Blockers**  
**Q**: What is a critical requirement for write blockers (per NIST)?  
**A**: Must prevent changes to protected drives while allowing read access.  

---

#### **14. File Carving**  
**Q**: How is file carving used in forensics?  
**A**: Recovers files from unallocated space using signatures (e.g., `‰PNG` for PNG files) when metadata is overwritten.  

---

#### **15. Key Tools**  
**Q**: Which tool is used for memory analysis in incident response?  
**A**: **Velociraptor** (or GRR/Falcon).  

---

#### **16. Cloud Evidence**  
**Q**: What type of data might cloud services provide?  
**A**: Social media logs, webmail backups, synced device data.  

---

#### **17. Number Systems**  
**Q**: Convert binary `0010 1011` to hex.  
**A**: `0x2B` (Left nibble: `0010` = 2, Right nibble: `1011` = B).  

---

#### **18. Chain of Custody**  
**Q**: Why is chain of custody crucial?  
**A**: Tracks evidence handling (who, when, purpose) to ensure legal admissibility.  

---

#### **19. Virtual Machine Artifacts**  
**Q**: How do VMs differ from physical systems in evidence collection?  
**A**: VMs store identical artifacts (OS/applications) but may be cloud-hosted or encrypted (e.g., Windows 11 VMs).  

---

#### **20. Forensic Workstations**  
**Q**: Name two features of a forensic workstation.  
**A**: High-end CPU/RAM, write blockers, validated tools (e.g., FTK Imager/X-Ways).  

---

### Summary  
These flashcards cover core DFIR concepts from evidence handling (ACPO principles, hashing) to technical processes (file carving, slack space analysis). Use them to reinforce key terms, tools, and workflows essential for digital investigations.  

**Source**: SANS DFIR Start Poster (2024).


Here is a clear, easy-to-read **essay** that summarizes and explains all the key content from the **SANS DFIR START** document, designed to give a comprehensive understanding of the Digital Forensics and Incident Response (DFIR) field:

---

# 🕵️‍♂️ Introduction to Digital Forensics and Incident Response (DFIR)

Digital Forensics and Incident Response (DFIR) is the cornerstone of investigating cyber incidents. It encompasses the tools, roles, processes, and methods needed to collect, analyze, and preserve digital evidence in a way that is both legally sound and technically robust.

---

## 🔍 What is DFIR?

At its core, **Digital Forensics** involves uncovering and interpreting electronic data to answer questions like who, what, when, where, why, and how. **Incident Response (IR)** is the structured approach to detecting, responding to, and recovering from security incidents.

---

## 📂 The Digital Forensics Process

1. **Acquisition**: Identify and collect data from a variety of sources (e.g., hard drives, memory).
2. **Examination & Analysis**: Interpret the data using specialized tools.
3. **Reporting**: Document findings in a clear, reproducible way.

This process is backed by international standards like ISO/IEC 27043.

---

## 🧭 The Incident Response Lifecycle

1. **Preparation**: Equip and train the team.
2. **Identification**: Detect incidents and assess scope.
3. **Containment**: Prevent further damage.
4. **Eradication**: Remove the cause of the incident.
5. **Recovery**: Restore systems and operations.
6. **Lessons Learned**: Review and improve.

---

## 💽 Understanding Data and Storage

* **File Systems**: Define how data is stored (e.g., NTFS, APFS, FAT32).
* **Sectors & Clusters**: Smallest units of disk storage.
* **Slack Space**: Unused data segments that may still hold forensic value.
* **RAM**: Volatile memory, rich in real-time data like passwords or clipboard contents.

---

## 🔢 Number Systems in Forensics

Understanding binary, hexadecimal, and decimal systems is vital for interpreting raw data:

* 8 bits = 1 byte = 2 nibbles
* 1 byte in binary (e.g., `00101011`) = `0x2B` in hex

---

## 📁 Digital Evidence & Metadata

* **Data**: The actual content (e.g., files).
* **Metadata**: Data about the data (e.g., file creation date, author).
* Tools like **ExifTool** help reveal hidden metadata.
* A file’s extension might lie—its **magic header** reveals its true type.

---

## 👨‍💼 Key DFIR Roles

| Role                      | Responsibilities                                           |
| ------------------------- | ---------------------------------------------------------- |
| **SOC Analyst**           | Monitors systems and detects threats                       |
| **First Responder**       | Identifies, preserves, and collects evidence               |
| **Forensic Investigator** | Acquires and extracts data                                 |
| **Forensic Analyst**      | Analyzes data and interprets what happened                 |
| **Incident Manager**      | Oversees operations, manages stress, and ensures resources |

---

## 🌍 Common Sources of Digital Evidence

* **Computers & Servers**: Contain OS and application artifacts.
* **Mobile Devices**: iOS and Android—often encrypted.
* **Cloud Services**: Store backups and logs.
* **IoT Devices & Drones**: May hold GPS, logs, or camera footage.
* **Vehicles**: Store logs, locations, and Bluetooth data.
* **Networks**: Provide packet-level evidence.
* **Industrial Systems**: SCADA and ICS environments.

---

## 🧰 DFIR Tools and Technologies

### 🧪 Acquisition Tools

Used to image or extract data from devices:

* FTK Imager
* Cellebrite UFED
* `dd` (Linux command)

### 🧬 Hex Editors

Let analysts view and edit files at the byte level:

* HxD
* WinHex
* 010 Editor

### 🧠 Forensic Suites

Used for deeper analysis, timeline reconstruction, and reporting:

* Magnet AXIOM
* Autopsy
* EnCase
* SANS SIFT Workstation

### 🛡️ Incident Response Tools

Monitor systems across a network:

* CrowdStrike Falcon
* Velociraptor
* GRR (Google Rapid Response)

### 🛠️ Specialized Tools

* Log2Timeline (timeline creation)
* USB Detective
* Arsenal Image Mounter

---

## 🖥️ Hardware & Field Equipment

* **Forensic Workstations**: High-performance computers built for analysis.
* **Write Blockers**: Prevent tampering with evidence.
* **Disk Duplicators**: Create exact copies of drives.
* **Network TAPs**: Capture network traffic.
* **Mobile Device Toolkits**: Connect and extract from smartphones/tablets.
* **Grab Bags**: Field kits with power cables, adapters, tools, and bootable USBs.

---

## 🧪 Key Concepts

### 🧾 Deleting a File

Deleted files may still be recoverable depending on what gets overwritten:

* **File carving**: Recover data without metadata.
* **Slack space**: Unused areas may hold traces of deleted data.

### 🔐 Hashing

Used to verify data integrity:

* MD5, SHA1, and SHA256 are common.
* Even a small change results in a drastically different hash.

### ⚖️ Locard’s Exchange Principle

Every interaction with a digital system leaves a trace—even anti-forensic actions can leave clues.

### 🧾 Chain of Custody

Tracks evidence handling throughout the process—crucial for legal admissibility.

### 🔄 Order of Volatility

Collect volatile data (RAM, CPU registers) **first**, then move to more stable sources (disk, backups).

---

## 📜 Foundational Principles

### 🛑 ACPO Principles

Followed globally in legal investigations:

1. Don’t alter original data.
2. If you access it, be able to justify your actions.
3. Maintain a verifiable audit trail.
4. Investigator holds ultimate responsibility for evidence integrity.

---

## 🧠 Final Thoughts

DFIR is a multidisciplinary, high-stakes field requiring technical skill, legal awareness, and methodical precision. From mobile devices to cloud systems and vehicles, evidence is everywhere—ready to be uncovered, preserved, and interpreted. Whether you're aspiring to be a forensic analyst, first responder, or SOC analyst, mastering DFIR means becoming a modern digital detective in a rapidly evolving landscape.

---

Would you like this turned into a downloadable PDF or a presentation-ready summ