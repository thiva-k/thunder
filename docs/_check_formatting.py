#!/usr/bin/env python3
"""Temporary script to check formatting issues in the threat model generator."""
import re

with open('generate_threat_model_docx.py', 'r') as f:
    lines = f.readlines()

print("=== Double spaces in string content ===")
count = 0
for i, line in enumerate(lines, 1):
    stripped = line.lstrip()
    # Only check strings (lines starting with quotes)
    if not ('"' in stripped):
        continue
    # Find double spaces between words inside the line
    matches = re.findall(r'[a-zA-Z]  [a-zA-Z]', line)
    if matches:
        count += 1
        print(f"  Line {i}: {matches[0]!r} in: {stripped[:100].strip()}")
print(f"  Total: {count}\n")

print("=== Inconsistent period usage in threat descriptions ===")
# Check if some threat entries end with period and some don't
threat_entries = []
for i, line in enumerate(lines, 1):
    # Threat table entries pattern: ("N","Category","Description...",
    m = re.search(r'"(\d+)","[^"]+","([^"]+)"', line)
    if m:
        num, desc = m.group(1), m.group(2)
        threat_entries.append((i, num, desc, desc.endswith('.')))

has_period = sum(1 for t in threat_entries if t[3])
no_period = sum(1 for t in threat_entries if not t[3])
print(f"  Threats ending with period: {has_period}")
print(f"  Threats NOT ending with period: {no_period}")
if no_period > 0 and has_period > 0:
    print("  INCONSISTENCY - entries without period:")
    for line_no, num, desc, has_p in threat_entries:
        if not has_p:
            print(f"    Line {line_no}: #{num}: ...{desc[-50:]}")
print()

print("=== Compliance table 'Section' prefix check ===")
in_compliance = False
for i, line in enumerate(lines, 1):
    if 'compliance_data' in line.lower() or 'rfc_9700_compliance' in line.lower():
        in_compliance = True
    if in_compliance and '"Section ' in line and '# ' not in line.lstrip()[:3]:
        print(f"  Line {i}: Still has 'Section' prefix: {line.strip()[:80]}")
    if in_compliance and line.strip() == ']':
        in_compliance = False
print()

print("=== Check for 'RFC Section' without number (missing RFC number) ===")
for i, line in enumerate(lines, 1):
    if 'RFC Section' in line and not re.search(r'RFC \d+ Section', line):
        print(f"  Line {i}: {line.strip()[:80]}")
print()

print("=== Check period consistency in mitigation descriptions ===")
mitigation_entries = []
for i, line in enumerate(lines, 1):
    # Mitigation is usually the last quoted string in threat tuples 
    # Pattern: "Yes/No","mitigation text")
    m = re.search(r'"(Yes|No)","([^"]+)"\)', line)
    if m:
        mit = m.group(2)
        mitigation_entries.append((i, mit, mit.endswith('.')))

has_period_m = sum(1 for t in mitigation_entries if t[2])
no_period_m = sum(1 for t in mitigation_entries if not t[2])
print(f"  Mitigations ending with period: {has_period_m}")
print(f"  Mitigations NOT ending with period: {no_period_m}")
if no_period_m > 0 and has_period_m > 0:
    print("  INCONSISTENCY - entries without period:")
    for line_no, mit, has_p in mitigation_entries:
        if not has_p:
            print(f"    Line {line_no}: ...{mit[-60:]}")
print()

print("=== Check for unescaped quotes or broken strings ===")
for i, line in enumerate(lines, 1):
    # Simple heuristic: odd number of unescaped quotes
    stripped = line.strip()
    if stripped.startswith('#'):
        continue
    quote_count = stripped.count('"') - stripped.count('\\"')
    if quote_count % 2 != 0 and not stripped.endswith('\\'):
        # Could be continuation, skip those
        if not (stripped.endswith('(') or stripped.endswith(',')):
            pass  # Too many false positives for continuation strings

print("=== Done ===")
