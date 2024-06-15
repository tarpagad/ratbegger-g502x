#!/usr/bin/env bash
# shellcheck disable=SC2086
set -h -u -o 'pipefail'

## SCRIPT INFORMATION
# ------------------------------------------------------------------------------
# Name          : RatBegger-G502X
# Description   : Load stored configuration files onto a Logitech G502 X mouse.
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

sScriptName='RatBegger-G502X'
sScriptVersion='20240615-1'
sRE_iRes='^resolution ([0-9])$'

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
    aMsg=()
    printf '\n'
    select sOp in 'Device Info' 'Load Profile' 'Set Active Profile' 'Help' 'Quit'
    do
        case "${sOp}" in
            ('Device Info') ratbagctl 'Logitech G502 X' info ;;
            ('Load Profile')
                printf '\n%s' 'Enter a profile number (0-4) to write to: '
                read -rN1 iProfile
                printf '\n\n%s\n\n' 'Select a configuration file to write...'
                select sFile in 'profiles/'*
                do
                    break
                done
                mapfile -t a < "${sFile}"
                for s in "${a[@]}" ; do
                    s1="$(cut -d '=' -f 1 <<< "${s}")"
                    s2="$(cut -d '=' -f 2 <<< "${s}")"
                    if [[ "${s1}" = ';'* ]] ; then
                        continue
                    elif [[ "${s1}" = 'profile name' ]] ; then # BUG https://github.com/libratbag/libratbag/issues/680
                        printf '%s\n' 'Seting profile name...'
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" name set "${s2}" || aMsg+=('Failed to set profile name.')
                    elif [[ "${s1}" = 'profile enable' ]] ; then
                        printf '%s\n' "Set profile '${s2}' state..."
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" "${s2}" || aMsg+=('Failed to enable/disable profile')
                    elif [[ "${s1}" = 'usb report rate' ]] ; then
                        printf '%s\n' 'Setting USB report rate...'
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" rate set "${s2}" || aMsg+=('Failed to set USB report rate.')
                    elif [[ "${s1}" = 'default resolution profile' ]] ; then
                        printf '%s\n' 'Setting default resolution profile...'
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" resolution default set "${s2}" || aMsg+=('Failed to set default resolution profile.')
                    elif [[ "${s1}" = 'active resolution profile' ]] ; then
                        printf '%s\n' 'Setting active resolution profile...'
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" resolution active set "${s2}" || aMsg+=('Failed to set active resolution profile.')
                    elif [[ "${s1}" =~ ${sRE_iRes} ]] ; then
                        printf '%s\n' "Setting resolution '${BASH_REMATCH[1]}'..."
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" resolution 0 dpi set "${BASH_REMATCH[1]}" || aMsg+=("Failed to set resolution ${BASH_REMATCH[1]}.")
                    elif [[ "${s2}" = 'button '* ]] ; then # NOTE don't quote ${s1} or ${s2}
                        printf '%s\n' "Mapping button '${s1}' to button '${s2}'..."
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" ${s1} action set ${s2} || aMsg+=("Failed to map button '${s1}' to button '${s2}'.")
                    elif [[ "${s2}" = *'KEY_'* ]] ; then # NOTE don't quote ${s1} or ${s2}
                        printf '%s\n' "Mapping button '${s1}' to key(s) '${s2}'..."
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" ${s1} action set macro ${s2} || aMsg+=("Failed to map button '${s1}' to key '${s2}'.")
                    elif [[ "${s2}" = *'-'* || "${s2}" = 'unknown' ]] ; then # assume 'special' action
                        # to list special actions, use ratbagctl with invalid action
                        printf '%s\n' "Mapping button '${s1}' to special action '${s2}'..."
                        ratbagctl 'Logitech G502 X' profile "${iProfile}" ${s1} action set special ${s2} || aMsg+=("Failed to map button '${s1}' to action '${s2}'.")
                    else
                        aMsg+=("Unknown configuration entry '${s}' in '${sFile}'.")
                    fi
                done

                if [[ -z "${aMsg[*]}" ]] ; then
                    printf '\n%s\n' "Profile '${iProfile}' written successfully!"
                else
                    printf '\n%s\n%s\n%s\n' "Profile '${iProfile}' write completed with errors:" '' "${aMsg[@]}"
                fi
            ;;
            ('Set Active Profile')
                read -rN1 -p 'Enter a profile number (0-4) to activate: ' iProfile
                printf '\n%s\n' "Setting active profile to '${iProfile}'..."
                ratbagctl 'Logitech G502 X' profile active set "${iProfile}" || printf '\n\n%s\n' "Failed to activate profile '${iProfile}'."
            ;;
            ('Help') xdg-open 'README.md' &> '/dev/null' && continue 2 ;;
            ('Quit') exit ;;
            (*) printf '%s\n' 'ERROR: Invalid choice!' && continue 2 ;;
        esac
        printf '\n%s\n' 'Press any key to continue.'
        read -rsN1
        continue 2
    done
done
