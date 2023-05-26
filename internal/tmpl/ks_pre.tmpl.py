#!/usr/bin/env python3
# This script is executed in Anaconda kickstart %pre section and then contents
# of /tmp/pre-generated.ks file is included and interpreted as the kickstart.
# Keep this Python 2 and Python 3 compatible.

from glob import glob
from pprint import pp
from subprocess import Popen, PIPE, run
import json
import os
import requests
from syslog import syslog

base_url = "{{ .BaseURL }}"
dmidecode = ["/usr/sbin/dmidecode"]
poweroff = ["/usr/sbin/poweroff", "--force", "--force"]
log = []

# executed this without Go templates: dev environment overrides
if base_url.startswith("{"):
    # connect to localhost rather than value from template
    base_url = "http://localhost:8000"
    # this needs to be in sudoers without password prompt in order to work
    dmidecode = ["sudo", "/usr/sbin/dmidecode"]
    # obviously we do not want to poweroff our machines during development
    poweroff = ["/usr/bin/true"]


def run_pipe(*cmd):
    try:
        proc = Popen(cmd, stdout=PIPE, stderr=PIPE)
        proc.wait()
        stdout = proc.stdout.read().decode().strip()
        stderr = proc.stderr.read().decode().strip()
        if len(stderr) > 0: log_write(" ".join(cmd), stderr)
        return stdout.strip()
    except FileNotFoundError:
        return ""


def log_write(prefix, message):
    global log
    log.append([prefix, str(message)])
    syslog(': '.join(["hardcap", prefix, str(message)]))


def ks_write(line):
    with open('/tmp/pre-generated.ks', 'a') as ks:
        ks.write(line)


def gather_mac():
    macs = []
    for name in glob("/sys/class/net/*/address"):
        mac = open(name).readline().strip()
        if len(mac) > 0 and mac != "00:00:00:00:00:00": macs.append(mac)
    return macs


def gather_serial():
    # also available via dmidecode but let's read it the same way as Anaconda does
    try:
        return open("/sys/class/dmi/id/product_serial").readline().strip()
    except Exception:
        return ""


def gather_facts():
    global log
    result = {
        "mac": gather_mac(),
        "serial": gather_serial(), # as in dracut/anaconda-ks-sendheaders.sh
        "cpu": {
            # TODO try with psutil package contains a lot of useful stuff
            "count": open('/proc/cpuinfo').read().count('processor\t:'),
        },
        "memory": {
            "bytes": os.sysconf('SC_PAGE_SIZE') * os.sysconf('SC_PHYS_PAGES'),
        },
        "dmi": {},
    }
    for keyword in ['bios-vendor', 'bios-version', 'bios-release-date', 'bios-revision', 'firmware-revision',
                    'system-manufacturer', 'system-product-name', 'system-version', 'system-serial-number',
                    'system-uuid', 'system-sku-number', 'system-family', 'baseboard-manufacturer',
                    'baseboard-product-name', 'baseboard-version', 'baseboard-serial-number', 'baseboard-asset-tag',
                    'chassis-manufacturer', 'chassis-type', 'chassis-version', 'chassis-serial-number',
                    'chassis-asset-tag', 'processor-family', 'processor-manufacturer', 'processor-version',
                    'processor-frequency']:
        result["dmi"][keyword] = run_pipe(*(dmidecode + ["-s", keyword]))
    result["log"] = log
    return result


facts = gather_facts()
print(json.dumps(facts, indent=2))

r = requests.post('%s/ks/register' % base_url, json=facts)
log_write("register upload", r.content)

# we are done, power off
ks_write("# will power off")
run(poweroff)
