$ErrorActionPreference = 'Stop'; # stop on all errors
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
. $toolsDir\common.ps1

try {
    uninstall_service
} catch {
    echo "$_"
}

try {
    remove_agent_registry_entries
} catch {
    echo "$_"
}
