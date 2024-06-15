#!/usr/bin/env bash
# shellcheck disable=SC2086
set -h -u -o 'pipefail'

## SCRIPT INFORMATION
# ------------------------------------------------------------------------------
# Name          : RatBagger-G502X
# License       : Non-Profit Open Software License ("Non-Profit OSL") 3.0
# Requirements  : GNU/Linux, Bash compatible shell
# Dependencies  : libratbag, mapfile
# Author        : 12bytes
# Website       : TODO
# Code          : https://codeberg.org/12bytes/ratbegger-g502x
# Credit        : Thanks to 'mercster' for prompting me to write this script
#                 (see https://github.com/libratbag/piper/issues/952).
# Usage         : See the included README.md file.
#
## LICENSE
# ------------------------------------------------------------------------------
# This program is free software as afforded by the Non-Profit Open Software
# License ("Non-Profit OSL") 3.0, a copy of which should have been included. The
# license can be viewed on-line at: https://opensource.org/licenses/NPOSL-3.0
# ------------------------------------------------------------------------------

sScriptName='RatBagger-G502X'
sScriptVersion='20240614'
aMsg=()

a=(
    ''
    '---------------------------'
    " ${sScriptName} v${sScriptVersion}"
    '          by 12bytes.org'
    '---------------------------'
    ''
)
printf '%s\n' "${a[@]}"

if ! type 'ratbagctl' &> '/dev/null' ; then
    printf '%s\n' 'ERROR: Missing libratbag dependency.' && exit 1
elif ! ratbagctl 'Logitech G502 X' info &> '/dev/null' ; then
    printf '%s\n' 'ERROR: Logitech G502 X not found.' && exit 1
fi
printf '%s\n' 'Found a Logitech G502 X'

while true ; do
    printf '\n'
    select sOp in 'Device Info' 'Load Profile' 'Set Active Profile' 'Help' 'Quit'
    do
        case "${sOp}" in
            ('Device Info') ratbagctl 'Logitech G502 X' info ;;
            ('Load Profile')
                printf '\n%s' 'Enter a profile number (0-4) to write to: '
                read -rN1 i
                printf '\n\n%s\n\n' 'Select a configuration file to write...'
                select sIniFile in 'profiles/'*
                do
                    break
                done
                mapfile -t a < "${sIniFile}"
                for s in "${a[@]}" ; do
                    s1="$(cut -d '=' -f 1 <<< "${s}")"
                    s2="$(cut -d '=' -f 2 <<< "${s}")"
                    if [[ "${s1}" = 'profile name' ]] ; then # BUG https://github.com/libratbag/libratbag/issues/680
                        printf '%s\n' 'Seting profile name...'
                        ratbagctl 'Logitech G502 X' profile "${i}" name set "${s2}" || aMsg+=('Failed to set profile name.')
                    elif [[ "${s1}" = 'profile enable' ]] ; then
                        printf '%s\n' "Set profile ${s2} state..."
                        ratbagctl 'Logitech G502 X' profile "${i}" "${s2}" || aMsg+=('Failed to enable/disable profile')
                    elif [[ "${s1}" = 'usb report rate' ]] ; then
                        printf '%s\n' 'Setting USB report rate...'
                        ratbagctl 'Logitech G502 X' profile "${i}" rate set "${s2}" || aMsg+=('Failed to set USB report rate.')
                    elif [[ "${s1}" = 'default resolution profile' ]] ; then
                        printf '%s\n' 'Setting default resolution profile...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution default set "${s2}" || aMsg+=('Failed to set default resolution profile.')
                    elif [[ "${s1}" = 'active resolution profile' ]] ; then
                        printf '%s\n' 'Setting active resolution profile...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution active set "${s2}" || aMsg+=('Failed to set active resolution profile.')
                    elif [[ "${s1}" = 'resolution 0' ]] ; then
                        printf '%s\n' 'Setting resolution 0...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution 0 dpi set "${s2}" || aMsg+=('Failed to set resolution 0.')
                    elif [[ "${s1}" = 'resolution 1' ]] ; then
                        printf '%s\n' 'Setting resolution 1...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution 1 dpi set "${s2}" || aMsg+=('Failed to set resolution 1.')
                    elif [[ "${s1}" = 'resolution 2' ]] ; then
                        printf '%s\n' 'Setting resolution 2...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution 2 dpi set "${s2}" || aMsg+=('Failed to set resolution 2.')
                    elif [[ "${s1}" = 'resolution 3' ]] ; then
                        printf '%s\n' 'Setting resolution 3...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution 3 dpi set "${s2}" || aMsg+=('Failed to set resolution 3.')
                    elif [[ "${s1}" = 'resolution 4' ]] ; then
                        printf '%s\n' 'Setting resolution 4...'
                        ratbagctl 'Logitech G502 X' profile "${i}" resolution 4 dpi set "${s2}" || aMsg+=('Failed to set resolution 4.')
                    elif [[ "${s2}" = 'button '* ]] ; then # NOTE don't quote ${s1} or ${s2}
                        printf '%s\n' "Mapping button ${s1} to button ${s2}..."
                        ratbagctl 'Logitech G502 X' profile "${i}" ${s1} action set ${s2} || aMsg+=("Failed to map button ${s1} to button ${s2}.")
                    elif [[ "${s1}" = 'button '* ]] ; then # NOTE don't quote ${s1} or ${s2}
                        printf '%s\n' "Mapping button ${s1} to key(s) ${s2}..."
                        ratbagctl 'Logitech G502 X' profile "${i}" ${s1} action set macro ${s2} || aMsg+=("Failed to map button ${s1} to key ${s2}.")
                    else # assume 'sepcial' action
                        printf '%s\n' "Mapping button ${s1} to special action ${s2}..."
                        ratbagctl 'Logitech G502 X' profile "${i}" ${s1} action set special ${s2} || aMsg+=("Failed to map button ${s1} to action ${s2}.")
                    fi
                done
            ;;
            ('Set Active Profile')
                read -rN1 -p 'Enter a profile number (0-4) to activate: ' i
                printf '%s\n' "Setting active profile to ${i}..."
                ratbagctl 'Logitech G502 X' profile active set "${i}" || aMsg+=("ERROR: Failed to activate profile ${i}.")
            ;;
            ('Help') xdg-open 'README.md' &> '/dev/null' ;;
            ('Quit') exit ;;
            (*) printf '%s\n' 'ERROR: Invalid choice!'
        esac

        if [[ -z "${aMsg[*]}" ]] ; then
            aMsg=('' 'Operation complete!')
        else
            aMsg=('' 'Operation completed with errors:' '' "${aMsg[@]}")
        fi
        aMsg+=('' 'Press any key to continue.')
        printf '%s\n' "${aMsg[@]}"
        read -rsN1
        continue 2
    done
done
