$ErrorActionPreference = 'Stop'

$toolsPath  = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

$packageArgs = @{
    PackageName  = "sclipi"
    File         = "$toolsPath\sclipi.zip"
    Destination  = $toolsPath
}
Get-ChocolateyUnzip @packageArgs

Remove-Item -force "$toolsPath\*.zip" -ea 0
