Great — I’ll analyze the **London Drugs** cyber incident using your required structure and also produce a 5-slide presentation outline you can use in class (with speaker notes and a short presenter script). I’ll cite reporting for the most important factual claims.

---

# Write-up

## Introduction

**What happened:** On **28 April 2024** London Drugs, a Western-Canada retail & pharmacy chain, discovered a cybersecurity incident (later claimed by the LockBit ransomware group). The company temporarily closed all 79 stores while investigating and restoring operations. LockBit publicly demanded a multi-million dollar ransom and later published some stolen corporate/employee data after negotiations failed. ([SC Media][1])

**Impact:** Stores were closed for days and some systems were offline (point-of-sale, phones). London Drugs reported it was investigating possible data theft; LockBit posted corp/employee files to prove exfiltration and demanded payment. Customers experienced disrupted services (including pharmacy access) and reputational damage followed. ([SC Media][1])

---

## Analysis

### CIA triad (Confidentiality, Integrity, Availability)

* **Confidentiality:** Likely breached — LockBit claims to have exfiltrated corporate and employee data and published some records. That indicates unauthorized data disclosure of corporate/employee information (confidentiality loss). ([Bitdefender][2])
* **Integrity:** Potential risk — ransomware often encrypts files, impacting integrity of data and systems; even if primary customer records were not confirmed compromised, the integrity of systems and backups may have been threatened during the incident. ([Itworldcanada][3])
* **Availability:** Clearly impacted — stores closed and services unavailable while containment and remediation occurred (availability loss). The company shut 79 stores until restoration. ([SC Media][1])

---

### Threat, Vulnerability, Threat-Action Scenarios

**Threat actors:** Ransomware criminal syndicate (LockBit claimed responsibility). Motivations: financial extortion and publicity. ([Bitdefender][2])

**Likely vulnerabilities exploited (typical for LockBit incidents):**

* Exposed/poorly patched edge services (RDP, VPN appliances) or weak credential protection.
* Insufficient segmentation allowing lateral movement from a compromised workstation to critical systems.
* Inadequate backup isolation (backups reachable from the infected network) or incomplete backup/restore testing.
  These are inferred from typical LockBit playbooks and observed operational impacts (store closures) — incident reports did not publicly disclose the exact initial vector. ([CM Alliance][4])

**Threat-action scenarios (plausible sequences):**

1. **Initial access:** Phishing or exposed remote access exploited → credential theft or exploit.
2. **Privilege escalation & lateral movement:** Attacker obtains domain credentials, moves to critical servers (POS, corporate file servers, backups).
3. **Exfiltration & encryption:** Sensitive files copied offsite (exfiltration) then systems encrypted; extortion demand issued.
4. **Disruption & publication:** Ransom refused/negotiated → attacker publishes stolen files to pressure payment. ([rdnewsnow.com][5])

---

### Affected IT infrastructure domains (typical mapping)

* **Perimeter & Remote Access:** VPN, remote management, firewalls (possible initial vector).
* **Endpoint & Workstation Security:** Laptops/desktops used by staff (phishing entry).
* **Identity & Access Management (IAM):** AD/Domain controllers, privileged accounts (path for lateral movement).
* **Server & Application Layer:** POS, corporate file servers, HR/Payroll systems (targeted for data and operational impact).
* **Backup & Recovery:** Backups must be isolated/immutable; if reachable, attacker can encrypt or delete backups.
* **Monitoring & Detection:** SIEM, EDR, logging — effectiveness determines time to detect and respond.

---

### Mitigation techniques & countermeasures

**Immediate / tactical (what organizations should do):**

* Isolate affected systems from the network to stop lateral spread.
* Engage incident response and forensics teams; preserve logs and evidence.
* Notify authorities / regulators and follow legal/notification requirements.
* Communicate transparently with customers and staff while avoiding operational panic. ([SC Media][1])

**Preventive / strategic (longer term):**

1. **Identity & Access**

   * Enforce MFA for all remote access and privileged accounts.
   * Use least privilege, break admin privileges into jump boxes/managed bastions.
2. **Edge hardening**

   * Patch and remove exposed RDP/VPN services; use up-to-date appliances.
3. **Endpoint & EDR**

   * Deploy modern Endpoint Detection & Response, with behavioral analytics to detect lateral movement and unusual exfiltration.
4. **Backups & Recovery**

   * Maintain air-gapped or immutable backups, test restores regularly, and keep backups offline/isolated from domain credentials.
5. **Network segmentation**

   * Microsegmentation so compromise in one segment (e.g., corporate office) can’t devastate POS or pharmacy systems.
6. **Logging & Monitoring**

   * Centralized logging, long retention, proactive threat hunting, and playbooks for alerts.
7. **Supply chain & third-party controls**

   * Review vendor access and limit third-party credentials; use MFA and least privilege for vendor accounts.
8. **Tabletop exercises & incident readiness**

   * Regular drills, contact lists for regulators, communications plans for customers and media.

**Countermeasure rationale:** These reduce attack surface, shorten detection time, make exfiltration harder, and ensure recoverability without paying ransom.

---

## Conclusion — Lessons learned & risk-management implications

1. **Availability risk to critical services:** Ransomware can force physical store closures and directly affect customers’ access to healthcare services (e.g., prescriptions). Business continuity planning for service interruptions must be prioritized. ([Click Armor][6])
2. **Data encryption + exfiltration = double extortion:** Organizations must assume exfiltration occurs; mitigation must include not only backups but also data minimization, encryption at rest, and threat detection to spot large outbound transfers. ([rdnewsnow.com][5])
3. **Trade-offs around ransom payment:** London Drugs reportedly refused a \$25M ransom and data was still published — illustrating that paying does not guarantee containment or reputation protection; decision frameworks and cyber insurance policies must be clear ahead of time. ([Global News][7])
4. **Holistic risk management:** Cybersecurity is not purely technical — it requires people, process, legal, communications, and emergency operations integration. Investment in detection, isolation, immutable backup, and readiness reduces business and reputational risk.
5. **Regulatory & customer trust impact:** Even if customer data is not confirmed leaked, public perception, regulatory scrutiny, and post-incident obligations (notification, remediation) create substantial downstream costs. ([SC Media][1])

---

# Class Presentation (≤5 slides) — content + speaker notes

Below are 5 slides with concise bullet content and speaker notes you can paste into PowerPoint/Google Slides.

---

## Slide 1 — Title & Overview

**Title:** London Drugs (Canada) — Ransomware Incident (Apr 2024)
**Bullets:**

* Retail & pharmacy chain — 79 stores temporarily closed.
* LockBit claimed attack; ransom demand \~\$25M.
* Data (corporate/employee) published after refusal. ([SC Media][1])

**Speaker notes (30–45s):**
Introduce the incident: date, scope (nationwide closures), attacker claim (LockBit), and immediate impact (store closures, potential data leak).

---

## Slide 2 — CIA Triad Impact

**Bullets:**

* Confidentiality: corporate/employee data exfiltrated & published. ([rdnewsnow.com][5])
* Integrity: potential encryption of systems & concerns about data tampering. ([Itworldcanada][3])
* Availability: full store closures; lost operations for >week. ([SC Media][1])

**Speaker notes (1min):**
Quickly map each triad dimension to concrete outcomes: public data release = confidentiality breach; encryption/lockout = integrity/availability issues; operational impact on customers.

---

## Slide 3 — Threat-Vulnerability-Action Scenarios

**Bullets:**

* Threat actor: LockBit (financial extortion). ([Bitdefender][2])
* Likely vulnerabilities: exposed remote access, weak MFA, insufficient segmentation.
* Attack flow summary: initial access → lateral movement → exfiltration → encryption → extortion/publish.

**Speaker notes (1:30):**
Explain plausible attack chain; emphasize typical LockBit tactics (phishing/RDP exploitation), and why segmentation/backups are crucial.

---

## Slide 4 — Affected Domains & Mitigations

**Left column (Domains):** Perimeter/Remote Access, Endpoints, IAM, Server/App, Backups, Monitoring.
**Right column (Key mitigations):** MFA & hardening, EDR + threat hunting, immutable/air-gapped backups, network segmentation, IR playbooks & tabletop exercises.

**Speaker notes (1:30):**
Walk through practical mitigations mapped to domains. Stress recovery readiness (tested restores) and proactive detection.

---

## Slide 5 — Conclusions & Lessons

**Bullets:**

* Ransomware = business continuity crisis (not just IT). ([Click Armor][6])
* Prepare for double extortion; assume exfiltration. ([rdnewsnow.com][5])
* Invest in detection, immutable backups, segmentation, and incident playbooks.
* Governance: clear decision paths for ransom, communication, and legal obligations.

**Speaker notes (45–60s):**
Summarize the big takeaways and what organizations should change immediately. End with a call for coordinated cyber risk management.

---

# Presentation logistics & leader responsibilities (for your class)

* **Group leader duties:** organize slides, open/close presentation, introduce team members, manage Q\&A. Ensure each slide presenter practices and that timing fits a 10-minute slot (rough guide: 30s intro, 3min CIA & scenario, 3min mitigations, 2min conclusions + 1min Q\&A).
* **Slide timing:** total 10 minutes; allocate \~2 minutes per slide (some slides a bit shorter).
* **Q\&A prep:** anticipate questions about ransom payment decision, regulatory notification requirements, and concrete steps to protect pharmacies/healthcare data.

---

# Sources (key citations)

* London Drugs store closures and incident discovery. ([SC Media][1])
* LockBit claimed responsibility and ransom demand reports. ([Bitdefender][2])
* Data published / stolen corporate files reporting. ([rdnewsnow.com][5])
* Operational & customer impacts analysis reporting. ([Click Armor][6])

---

If you want, I can:

* produce the actual PowerPoint file (.pptx) with the 5 slides + speaker notes, ready to present, or
* expand any section (e.g., provide a detailed incident response checklist, a timeline of the London Drugs attack from public sources, or mock Q\&A answers for class).

Which would you like next?

[1]: https://www.scworld.com/news/london-drugs-pharmacy-closes-all-stores-to-respond-to-cyber-incident?utm_source=chatgpt.com "London Drugs pharmacy closes all stores to respond ..."
[2]: https://www.bitdefender.com/en-us/blog/hotforsecurity/lockbit-claims-recent-attack-on-london-drugs-pharmacy-chain?utm_source=chatgpt.com "LockBit Claims Recent Attack on London Drugs Pharmacy ..."
[3]: https://www.itworldcanada.com/article/cyber-security-today-may-1-2024-data-may-have-been-stolen-in-london-drugs-cyber-attack-congressional-testimony-today-by-unitedhealth-ceo-on-ransomware-attack-and-more/560716?utm_source=chatgpt.com "Data may have been stolen in London Drugs cyber attack ..."
[4]: https://www.cm-alliance.com/cybersecurity-blog/may-2024-biggest-cyber-attacks-data-breaches-ransomware-attacks?utm_source=chatgpt.com "May 2024: Biggest Cyber Attacks, Data Breaches & ..."
[5]: https://rdnewsnow.com/2024/05/23/hackers-release-corporate-data-stolen-from-london-drugs-company-says/?utm_source=chatgpt.com "Hackers release corporate data stolen from London Drugs ..."
[6]: https://clickarmor.ca/london-drugs-cyber-attack-what-businesses-can-learn-from-its-week-long-shutdown/?utm_source=chatgpt.com "London Drugs cyber attack: What businesses can learn"
[7]: https://globalnews.ca/news/10516121/london-drugs-ransom-attack-employee/?utm_source=chatgpt.com "London Drugs hackers seek millions in ransom on claims ..."
