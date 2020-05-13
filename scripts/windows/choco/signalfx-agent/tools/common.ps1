$program_data_path = "\ProgramData\SignalFxAgent"
$installation_path = "\Program Files"
$config_path = "$program_data_path\agent.yaml"

function get_value_from_file([string]$path) {
    $value = ""
    if (Test-Path -Path "$path") {
        $value = (Get-Content -Path "$path").Trim()
    }
    return "$value"
}

# create directories in program data
function create_program_data() {
    mkdir "$program_data_path" -ErrorAction Ignore
}

# whether the agent service is running
function service_running() {
    return (((Get-CimInstance -ClassName win32_service -Filter "Name = 'SignalFx Smart Agent'" | Select Name, State).State -Eq "Running") -Or ((Get-CimInstance -ClassName win32_service -Filter "Name = 'signalfx-agent'" | Select Name, State).State -Eq "Running"))
}

# whether the agent service is installed
function service_installed() {
    return (((Get-CimInstance -ClassName win32_service -Filter "Name = 'SignalFx Smart Agent'" | Select Name, State).Name -Eq "SignalFx Smart Agent") -Or ((Get-CimInstance -ClassName win32_service -Filter "Name = 'signalfx-agent'" | Select Name, State).Name -Eq "signalfx-agent"))
}

# start the service if it's stopped
function start_service([string]$installation_path=$installation_path, [string]$config_path=$config_path) {
    if (!(service_running)){
        $agent_bin = Resolve-Path "$installation_path\SignalFx\SignalFxAgent\bin\signalfx-agent.exe"
        & $agent_bin -service "start" -config "$config_path"
    }
}

# stop the service if it's running
function stop_service([string]$installation_path=$installation_path, [string]$config_path=$config_path) {
    if ((service_running)){
        $agent_bin = Resolve-Path "$installation_path\SignalFx\SignalFxAgent\bin\signalfx-agent.exe"
        & $agent_bin -service "stop" -config "$config_path"
    }
}

# remove registry entries created by the agent service
function remove_agent_registry_entries() {
    try
    {
        if (Test-Path "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\SignalFx Smart Agent"){
            Remove-Item "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\SignalFx Smart Agent"
        }
        if (Test-Path "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\signalfx-agent"){
            Remove-Item "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\signalfx-agent"
        }
    } catch {
        $err = $_.Exception.Message
        $message = "
        unable to remove registry entries at HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\SignalFx Smart Agent
        $err
        "
        throw "$message"
    }
}

# install the service if it's not already installed
function install_service([string]$installation_path=$installation_path, [string]$config_path=$config_path) {
    if (!(service_installed)){
        $agent_bin = Resolve-Path "$installation_path\SignalFx\SignalFxAgent\bin\signalfx-agent.exe"
        & $agent_bin -service "install" -logEvents -config "$config_path"
    }
}

# uninstall the service
function uninstall_service([string]$installation_path=$installation_path) {
    if ((service_installed)){
        stop_service -installation_path $installation_path -config_path $config_path
        $agent_bin = Resolve-Path "$installation_path\SignalFx\SignalFxAgent\bin\signalfx-agent.exe"
        & $agent_bin -service "uninstall" -logEvents
    }
}

# check registry for the agent msi package
function msi_installed([string]$name="SignalFx Smart Agent") {
    return (Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\* | Where { $_.DisplayName -eq $name }) -ne $null
}
