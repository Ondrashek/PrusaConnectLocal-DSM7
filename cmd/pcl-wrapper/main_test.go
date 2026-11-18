package main
import("os";"path/filepath";"testing")
func TestDSMPorts(t *testing.T){p:=filepath.Join(t.TempDir(),"synoinfo.conf"); _=os.WriteFile(p,[]byte("admin_port=\"5100\"\nsecure_admin_port=\"5101\"\nother=\"5001\"\n"),0600); v:=dsmPorts(p);if v["admin_port"]!=5100||v["secure_admin_port"]!=5101||len(v)!=2{t.Fatal(v)}}
func TestCode(t *testing.T){v:=codePattern.FindStringSubmatch(`level=WARN setup_code=X4ZF-VBX9-GMPH`);if len(v)!=2||v[1]!="X4ZF-VBX9-GMPH"{t.Fatal(v)}}
func TestDigest(t *testing.T){if digest("a","b")==digest("c","b"){t.Fatal("salt ignored")}}
