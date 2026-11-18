package main

import (
 "bytes"
 "crypto/sha256"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "gopkg.in/yaml.v3"
)

// YAML is parsed, never patched with regex. Original bytes are backed up before
// the first change. Unrelated mapping values (database, accounts, storage) survive.
func mapping(parent map[string]any, key string) (map[string]any,error) {
 v,ok:=parent[key]
 if !ok || v==nil { m:=map[string]any{}; parent[key]=m; return m,nil }
 m,ok:=v.(map[string]any); if !ok { return nil,fmt.Errorf("%s must be a mapping",key) }
 return m,nil
}
func configForDirectHTTPS(old []byte, dir string) ([]byte,error) {
 doc:=map[string]any{}
 if len(bytes.TrimSpace(old))>0 {
  dec:=yaml.NewDecoder(bytes.NewReader(old))
  if err:=dec.Decode(&doc); err!=nil { return nil,fmt.Errorf("invalid YAML") }
  if doc==nil { doc=map[string]any{} }
  var extra any
  if err:=dec.Decode(&extra); err!=io.EOF { return nil,fmt.Errorf("expected one YAML document") }
 }
 core,err:=mapping(doc,"core"); if err!=nil { return nil,err }
 api,err:=mapping(core,"api"); if err!=nil { return nil,err }
 api["addr"]=fmt.Sprintf(":%d",httpsPort)
 tls,err:=mapping(api,"tls"); if err!=nil { return nil,err }
 tls["certificate"]=filepath.Join(dir,"https.crt")
 tls["key"]=filepath.Join(dir,"https.key")
 upgrade,err:=mapping(core,"upgrade"); if err!=nil { return nil,err }; upgrade["enabled"]=false
 // Do not change an explicitly configured connector or its printer settings.
 // Existing wizard connector settings live in config.d/setup.yaml. If present,
 // preserve them and let the upstream merge them below this file.
 setupPath:=filepath.Join(dir,"config","config.d","setup.yaml")
 hasWizardConnectors:=false
 if b,e:=os.ReadFile(setupPath); e==nil {
  var setup map[string]any
  if e=yaml.Unmarshal(b,&setup); e!=nil { return nil,fmt.Errorf("invalid wizard YAML") }
  if c,ok:=setup["core"].(map[string]any); ok { _,hasWizardConnectors=c["connectors"] }
 } else if !os.IsNotExist(e) { return nil,e }
 if _,ok:=core["connectors"]; !ok && !hasWizardConnectors {
  core["connectors"]=map[string]any{"xbuddy":map[string]any{"type":"xbuddy","config":map[string]any{"listener":map[string]any{"addr":fmt.Sprintf(":%d",backendPort)}}}}
 }
 return yaml.Marshal(doc)
}
func migrateConfig(dir string) error {
 folder:=filepath.Join(dir,"config"); if err:=os.MkdirAll(folder,0700); err!=nil { return err }
 path:=filepath.Join(folder,"config.yaml")
 if info,e:=os.Lstat(path); e==nil && !info.Mode().IsRegular() { return fmt.Errorf("config.yaml must be a regular file") }
 old,err:=os.ReadFile(path); existed:=err==nil
 if err!=nil && !os.IsNotExist(err) { return err }
 next,err:=configForDirectHTTPS(old,dir); if err!=nil { return err }
 if bytes.Equal(old,next) { return nil }
 if existed {
  hash:=sha256.Sum256(old)
  backup:=filepath.Join(folder,fmt.Sprintf("config.yaml.before-v5-%x.bak",hash[:8]))
  f,e:=os.OpenFile(backup,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
  if e==nil { _,e=f.Write(old); if e==nil { e=f.Sync() }; ce:=f.Close(); if e!=nil { return e }; if ce!=nil { return ce }
  } else if os.IsExist(e) { b,re:=os.ReadFile(backup); if re!=nil || !bytes.Equal(b,old) { return fmt.Errorf("backup mismatch") }
  } else { return e }
 }
 f,err:=os.CreateTemp(folder,".config-v5-*"); if err!=nil { return err }
 tmp:=f.Name(); defer os.Remove(tmp)
 if err=f.Chmod(0600); err==nil { _,err=f.Write(next) }
 if err==nil { err=f.Sync() }; ce:=f.Close()
 if err!=nil { return err }; if ce!=nil { return ce }
 return os.Rename(tmp,path)
}
