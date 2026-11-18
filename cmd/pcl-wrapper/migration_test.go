package main
import("os";"path/filepath";"testing";"bytes";"gopkg.in/yaml.v3")
func TestDirectHTTPSMigration(t *testing.T){
 dir:=t.TempDir(); folder:=filepath.Join(dir,"config"); _=os.MkdirAll(folder,0700)
 old:=[]byte("core:\n  api:\n    addr: ':18080'\n    request_timeout: 45s\npersistence:\n  type: sqlite\nresources:\n  storage:\n    local:\n      type: filesystem\n")
 path:=filepath.Join(folder,"config.yaml"); if err:=os.WriteFile(path,old,0600);err!=nil{t.Fatal(err)}
 if err:=migrateConfig(dir);err!=nil{t.Fatal(err)}
 b,_:=os.ReadFile(path);var doc map[string]any;if err:=yaml.Unmarshal(b,&doc);err!=nil{t.Fatal(err)}
 core:=doc["core"].(map[string]any);api:=core["api"].(map[string]any)
 if api["addr"]!=":18443"||api["request_timeout"]!="45s"{t.Fatal(api)}
 if doc["persistence"].(map[string]any)["type"]!="sqlite"||doc["resources"]==nil{t.Fatal("lost settings")}
 tls:=api["tls"].(map[string]any);if tls["certificate"]!=filepath.Join(dir,"https.crt"){t.Fatal(tls)}
 backups,_:=filepath.Glob(filepath.Join(folder,"*.bak"));if len(backups)!=1{t.Fatal(backups)}
 backup,_:=os.ReadFile(backups[0]);if !bytes.Equal(old,backup){t.Fatal("backup differs")}
 if err:=migrateConfig(dir);err!=nil{t.Fatal(err)};again,_:=os.ReadFile(path);if !bytes.Equal(b,again){t.Fatal("not idempotent")}
}
func TestRejectMalformed(t *testing.T){for _,b:=range []string{"core: [", "core: 1", "core: {}\n---\ncore: {}"}{if _,err:=configForDirectHTTPS([]byte(b),t.TempDir());err==nil{t.Fatal("accepted",b)}}}
func TestPreserveWizardConnector(t *testing.T){
 dir:=t.TempDir();p:=filepath.Join(dir,"config","config.d");_=os.MkdirAll(p,0700)
 _=os.WriteFile(filepath.Join(p,"setup.yaml"),[]byte("core:\n  connectors:\n    xbuddy:\n      type: xbuddy\n"),0600)
 b,err:=configForDirectHTTPS(nil,dir);if err!=nil{t.Fatal(err)};var d map[string]any;_=yaml.Unmarshal(b,&d)
 if _,ok:=d["core"].(map[string]any)["connectors"];ok{t.Fatal("overrode wizard connectors")}
}
