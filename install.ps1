$googet_root = "$env:ProgramData\GooGet"

$root = [Environment]::GetEnvironmentVariable('GooGetRoot', 'Machine')
if ($root -ne "$googet_root") {
  [Environment]::SetEnvironmentVariable('GooGetRoot', "$googet_root", 'Machine')
}

$path = [Environment]::GetEnvironmentVariable('Path', 'Machine')
if ($path -notlike "*%GooGetRoot%*") {
  $path = $path + ";%GooGetRoot%"
  [Environment]::SetEnvironmentVariable('Path', $path, 'Machine')
}
