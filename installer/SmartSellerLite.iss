#define AppName "SmartSeller Lite"
#define AppExe  "SmartSellerLite.exe"
#define AppVer  "0.1.0"
#define AppDir  "{commonpf}\SmartSeller Lite"
#define IconFile AddBackslash(SourcePath) + "icon.ico"

#pragma message ">>> Using icon: " + IconFile
#ifnexist IconFile
  #error "icon.ico NOT FOUND at: " + IconFile
#endif

[Setup]
AppName={#AppName}
AppPublisher=bie7
AppVersion={#AppVer}
DefaultDirName={#AppDir}
DefaultGroupName={#AppName}
OutputBaseFilename=SmartSellerLite-Setup
Compression=lzma
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
SetupIconFile={#IconFile}

[Dirs]
Name: "{commonappdata}\SmartSellerLite"; Permissions: users-modify
Name: "{commonappdata}\SmartSellerLite\config"; Permissions: users-modify

[Files]
; Aplikasi utama
Source: "{#SourcePath}\..\build\{#AppExe}"; DestDir: "{app}"; Flags: ignoreversion
; Icon untuk shortcut
Source: "{#IconFile}"; DestDir: "{app}"; Flags: ignoreversion
; MariaDB MSI (diextract saat dibutuhkan)
Source: "{#SourcePath}\thirdparty\mariadb.msi"; Flags: dontcopy

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExe}"; IconFilename: "{app}\icon.ico"
Name: "{commondesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; IconFilename: "{app}\icon.ico"

[UninstallRun]
Filename: "sc.exe"; Parameters: "stop MariaDB"; Flags: runhidden; RunOnceId: "StopMariaDBGeneric"

[Code]
var
  PgDB: TWizardPage;
  RBUseExisting, RBInstallBundled: TRadioButton;
  LExample, LLocalInfo, LDSN, LRootInfo: TNewStaticText;
  EdDSN, EdRootPass: TNewEdit;
  InstallMariaDB: Boolean;
  RootPassForBundled: string;
  SelectedDSN: string;

function TrimS(const S: string): string;
begin
  Result := Trim(S);
end;

function RunAndWait(const FileName, Params: string; Hide: Boolean): Integer;
var
  ShowCmd: Integer;
  ExecOk: Boolean;
  ExitCode: Integer;
begin
  if Hide then
    ShowCmd := SW_HIDE
  else
    ShowCmd := SW_SHOWNORMAL;

  ExecOk := Exec(FileName, Params, '', ShowCmd, ewWaitUntilTerminated, ExitCode);
  if ExecOk then
    Result := ExitCode
  else
    Result := -1;
end;
function DirExistsMask(const Mask: string): string;
var
  FR: TFindRec;
begin
  Result := '';
  if FindFirst(Mask, FR) then
  try
    repeat
      if (FR.Attributes and FILE_ATTRIBUTE_DIRECTORY) <> 0 then
      begin
        Result := FR.Name;
        Exit;
      end;
    until not FindNext(FR);
  finally
    FindClose(FR);
  end;
end;

function FindMariaDBBinDir(): string;
var
  base, found, candidate: string;
begin
  base := ExpandConstant('{commonpf}');
  found := DirExistsMask(base + '\MariaDB*');
  if found <> '' then
  begin
    candidate := base + '\' + found + '\bin';
    if DirExists(candidate) and FileExists(AddBackslash(candidate) + 'mysql.exe') then
    begin
      Result := candidate;
      Exit;
    end;
  end;

  base := ExpandConstant('{commonpf32}');
  found := DirExistsMask(base + '\MariaDB*');
  if found <> '' then
  begin
    candidate := base + '\' + found + '\bin';
    if DirExists(candidate) and FileExists(AddBackslash(candidate) + 'mysql.exe') then
    begin
      Result := candidate;
      Exit;
    end;
  end;

  Result := '';
end;

function WaitPortReady(const Host: string; Port, Attempts, MsInterval: Integer): Boolean;
var
  Cmd: string;
  Rc: Integer;
begin
  Cmd :=
    '-NoProfile -Command "for($i=0;$i -lt ' + IntToStr(Attempts) + ';' +
    '$i++){try{$c=New-Object Net.Sockets.TcpClient;$c.Connect(''' + Host + ''',' + IntToStr(Port) + ');$c.Close();exit 0}' +
    'catch{Start-Sleep -Milliseconds ' + IntToStr(MsInterval) + '}} exit 1"';
  Rc := 1;
  Exec('powershell.exe', Cmd, '', SW_HIDE, ewWaitUntilTerminated, Rc);
  Result := Rc = 0;
end;





procedure CreateSmartsellerDB(const BinDir, Host: string; Port: Integer; const RootPass: string);
var
  Rc: Integer;
  MysqlExe, Sql: string;
begin
  MysqlExe := AddBackslash(BinDir) + 'mysql.exe';
  Sql := 'CREATE DATABASE IF NOT EXISTS `smartseller` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;';
  if not Exec('cmd.exe',
    '/C "' + MysqlExe + '" -h ' + Host + ' -P ' + IntToStr(Port) + ' -uroot -p' + RootPass + ' -e "' + Sql + '"',
    '', SW_HIDE, ewWaitUntilTerminated, Rc) then
    MsgBox('Gagal menjalankan mysql.exe untuk membuat database. Cek instalasi MariaDB.', mbError, MB_OK)
  else if Rc <> 0 then
    MsgBox('Create database gagal (kode ' + IntToStr(Rc) + '). Periksa password root atau port.', mbError, MB_OK);
end;


procedure UpdateDBMode(Sender: TObject);
begin
  EdRootPass.Enabled := RBInstallBundled.Checked;
  EdDSN.Enabled := RBUseExisting.Checked;
end;

procedure InitializeWizard;
begin
  PgDB := CreateCustomPage(wpSelectDir, 'Database Setup', 'Pilih cara menyiapkan database untuk SmartSeller.');

  RBUseExisting := TRadioButton.Create(PgDB.Surface);
  RBUseExisting.Parent := PgDB.Surface;
  RBUseExisting.Left := ScaleX(8);
  RBUseExisting.Top := ScaleY(8);
  RBUseExisting.Width := ScaleX(520);
  RBUseExisting.Caption := 'Gunakan database yang sudah ada (masukkan DSN manual)';
  RBUseExisting.Checked := True;
  RBUseExisting.OnClick := @UpdateDBMode;

  LDSN := TNewStaticText.Create(PgDB.Surface);
  LDSN.Parent := PgDB.Surface;
  LDSN.Left := ScaleX(24);
  LDSN.Top := ScaleY(36);
  LDSN.Width := ScaleX(520);
  LDSN.Caption := 'DSN:';

  EdDSN := TNewEdit.Create(PgDB.Surface);
  EdDSN.Parent := PgDB.Surface;
  EdDSN.Left := ScaleX(24);
  EdDSN.Top := ScaleY(54);
  EdDSN.Width := ScaleX(520);
  EdDSN.Text := 'user:pass@tcp(127.0.0.1:3306)/smartseller?parseTime=true&charset=utf8mb4&loc=Local';

  LExample := TNewStaticText.Create(PgDB.Surface);
  LExample.Parent := PgDB.Surface;
  LExample.Left := ScaleX(24);
  LExample.Top := ScaleY(80);
  LExample.Width := ScaleX(520);
  LExample.Caption := 'Contoh format DSN: user:pass@tcp(127.0.0.1:3306)/smartseller?parseTime=true&charset=utf8mb4&loc=Local';

  RBInstallBundled := TRadioButton.Create(PgDB.Surface);
  RBInstallBundled.Parent := PgDB.Surface;
  RBInstallBundled.Left := ScaleX(8);
  RBInstallBundled.Top := ScaleY(112);
  RBInstallBundled.Width := ScaleX(520);
  RBInstallBundled.Caption := 'Install MariaDB dari paket ini (menjalankan MariaDB.msi)';
  RBInstallBundled.OnClick := @UpdateDBMode;
  LLocalInfo := TNewStaticText.Create(PgDB.Surface);
  LLocalInfo.Parent := PgDB.Surface;
  LLocalInfo.Left := ScaleX(24);
  LLocalInfo.Top := ScaleY(140);
  LLocalInfo.Width := ScaleX(520);
  LLocalInfo.Caption := 'MariaDB.msi akan muncul dengan wizard resminya. Selesaikan instalasi tersebut lalu kembali ke installer ini.';

  LRootInfo := TNewStaticText.Create(PgDB.Surface);
  LRootInfo.Parent := PgDB.Surface;
  LRootInfo.Left := ScaleX(24);
  LRootInfo.Top := ScaleY(168);
  LRootInfo.Width := ScaleX(520);
  LRootInfo.Caption := 'Masukkan password root yang sama dengan yang Anda set saat wizard MariaDB.msi.';

  EdRootPass := TNewEdit.Create(PgDB.Surface);
  EdRootPass.Parent := PgDB.Surface;
  EdRootPass.Left := ScaleX(24);
  EdRootPass.Top := ScaleY(188);
  EdRootPass.Width := ScaleX(520);
  EdRootPass.PasswordChar := '*';
  InstallMariaDB := False;
  SelectedDSN := '';
  RootPassForBundled := '';
  UpdateDBMode(nil);
end;

function NextButtonClick(CurPageID: Integer): Boolean;
var
  EnvStr: string;
begin
  Result := True;
  if CurPageID <> PgDB.ID then
    Exit;

  InstallMariaDB := RBInstallBundled.Checked;

  if RBUseExisting.Checked then
  begin
    SelectedDSN := TrimS(EdDSN.Text);
    if SelectedDSN = '' then
    begin
      MsgBox('DSN tidak boleh kosong.', mbError, MB_OK);
      Result := False;
      Exit;
    end;
    RootPassForBundled := '';
  end
  else
  begin
    RootPassForBundled := TrimS(EdRootPass.Text);
    if RootPassForBundled = '' then
    begin
      MsgBox('Password root wajib diisi untuk melanjutkan instalasi MariaDB bundel.', mbError, MB_OK);
      Result := False;
      Exit;
    end;
    SelectedDSN := 'root:' + RootPassForBundled + '@tcp(127.0.0.1:3306)/smartseller?parseTime=true&charset=utf8mb4&loc=Local';
  end;
  EnvStr :=
    'APP_BRAND_NAME=SmartSeller Lite'#13#10 +
    'APP_ADDR=127.0.0.1:8787'#13#10 +
    'APP_OPEN_BROWSER=true'#13#10 +
    'DB_DRIVER=mysql'#13#10 +
    'DATABASE_DSN=' + SelectedDSN + #13#10;

  ForceDirectories(ExpandConstant('{commonappdata}\SmartSellerLite\config'));
  SaveStringToFile(ExpandConstant('{commonappdata}\SmartSellerLite\config\.env'), EnvStr, False);
  SaveStringToFile(ExpandConstant('{app}\.env'), EnvStr, False);
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ExecRes: Integer;
  BinDir: string;
begin
  if CurStep <> ssPostInstall then
    Exit;

  ExecRes := 0;
  if InstallMariaDB then
  begin
    ExtractTemporaryFile('mariadb.msi');
    ExecRes := RunAndWait('msiexec.exe', '/i "' + ExpandConstant('{tmp}\mariadb.msi') + '"', False);
    if (ExecRes <> 0) and (ExecRes <> 3010) then
      MsgBox('Instalasi MariaDB selesai dengan kode ' + IntToStr(ExecRes) + '. Provisioning database tetap dicoba.', mbInformation, MB_OK);

    BinDir := FindMariaDBBinDir();
    if BinDir = '' then
      MsgBox('Tidak bisa menemukan MariaDB\bin. Pastikan instalasi MSI sukses.', mbError, MB_OK)
    else
    begin
      if not WaitPortReady('127.0.0.1', 3306, 40, 500) then
        MsgBox('MariaDB belum listen di port 3306. Provisioning database mungkin gagal.', mbInformation, MB_OK)
      else if RootPassForBundled <> '' then
        CreateSmartsellerDB(BinDir, '127.0.0.1', 3306, RootPassForBundled);
    end;
  end;

  ShellExec('', ExpandConstant('{app}\{#AppExe}'), '', ExpandConstant('{app}'), SW_SHOWNORMAL, ewNoWait, ExecRes);
  ExecRes := RunAndWait('powershell.exe',
    '-NoProfile -Command "for($i=0;$i -lt 40;$i++){try{$c=New-Object Net.Sockets.TcpClient;$c.Connect(''127.0.0.1'',8787);$c.Close();exit 0}catch{Start-Sleep -Milliseconds 500}} exit 1"',
    True);
  if ExecRes = 0 then
    ShellExec('', 'http://127.0.0.1:8787', '', '', SW_SHOWNORMAL, ewNoWait, ExecRes);
end;















