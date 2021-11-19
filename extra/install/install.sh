#!/bin/bash
#
# Installation script for labe tools. Mostly copies static binaries in place,
# because there is not much else possible on the vanilla target.

set -eu -o pipefail

# Where to install binaries.
BIN=${HOME}/.local/bin
# Name of uninstall script.
UNINSTALL_SCRIPT_NAME="labe-uninstall.sh"
# Programs to install.
PROGS=(
    "labed"
    "makta"
    "solrdump"
    "sqlite3"
    "tabbedjson"
)

# ================ ================

mkdir -p "$BIN"
UNINSTALL_SCRIPT=${BIN}/${UNINSTALL_SCRIPT_NAME}

cat <<EOM

LABE INSTALLER

To uninstall, run:

    \$ ${UNINSTALL_SCRIPT_NAME}

Installing utilities:

EOM

for prog in ${PROGS[@]}; do
    cp -v "$prog" "$BIN"
done

cat <<EOM > ${UNINSTALL_SCRIPT}
#!/bin/bash
#
# Uninstaller for labe, autogenerated on $(date)

set -eu -o pipefail
cd ${BIN} && rm -f ${PROGS[@]} && cd "$OLDPWD"
rm -- ${UNINSTALL_SCRIPT}
echo "[ok] labe uninstalled successfully"
EOM
chmod +x ${UNINSTALL_SCRIPT}
