// Unofficial Synology integration. See README.md before deployment.
package main

import (
 "bufio"
 "context"
 "crypto/ecdsa"
 "crypto/pbkdf2"
 "crypto/elliptic"
 "crypto/rand"
 "crypto/sha256"
 "crypto/subtle"
 "crypto/tls"
 "crypto/x509"
 "crypto/x509/pkix"
 "encoding/hex"
 "encoding/json"
 "encoding/pem"
 "fmt"
 "html/template"
 "log"
 "math/big"
 "net"
 "net/http"
 "net/http/httputil"
 "net/url"
 "os"
 "os/exec"
 "os/signal"
 "path/filepath"
 "regexp"
 "strconv"
 "strings"
 "sync"
 "syscall"
 "time"
)
const httpsPort = 18443
const backendPort = 18080
var state string
var mu sync.RWMutex
var setupCode string
var setupAt time.Time
var childStatus = "starting"
var codePattern = regexp.MustCompile(`setup_code="?([A-Z0-9]+-[A-Z0-9]+-[A-Z0-9]+)`)
type credential struct { Salt string; Digest string }
func must(err error) { if err != nil { log.Fatal(err) } }
func digest(salt, password string) string { h,err:=pbkdf2.Key(sha256.New,password,[]byte(salt),600000,32); if err!=nil { panic(err) }; return hex.EncodeToString(h) }
func initCredential() {
 p:=os.Getenv("wizard_portal_password")
 if len(p)<8 || len(p)>256 { log.Fatal("Portal password must contain 8 to 256 bytes") }
 b:=make([]byte,32); _,err:=rand.Read(b); must(err)
 c:=credential{Salt:hex.EncodeToString(b)}; c.Digest=digest(c.Salt,p)
 data,err:=json.Marshal(c); must(err)
 must(os.WriteFile(filepath.Join(state,"portal-auth.json"),data,0600))
}
func auth(next http.Handler) http.Handler {
 var c credential
 b,err:=os.ReadFile(filepath.Join(state,"portal-auth.json")); must(err); must(json.Unmarshal(b,&c))
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  u,p,ok:=r.BasicAuth()
  if !ok || u!="admin" || subtle.ConstantTimeCompare([]byte(digest(c.Salt,p)),[]byte(c.Digest))!=1 {
   // Delay adds friction; this is not a complete brute-force lockout mechanism.
   time.Sleep(700*time.Millisecond)
   w.Header().Set("WWW-Authenticate",`Basic realm="Prusa Connect setup", charset="UTF-8"`)
   http.Error(w,"Authentication required",401); return
  }
  next.ServeHTTP(w,r)
 })
}
func certFiles() (string,string) {
 cert:=filepath.Join(state,"https.crt"); key:=filepath.Join(state,"https.key")
 if _,err:=tls.LoadX509KeyPair(cert,key); err==nil { return cert,key }
 k,err:=ecdsa.GenerateKey(elliptic.P256(),rand.Reader); must(err)
 serial,err:=rand.Int(rand.Reader,new(big.Int).Lsh(big.NewInt(1),128)); must(err)
 host,_:=os.Hostname()
 t:=x509.Certificate{SerialNumber:serial,Subject:pkix.Name{CommonName:host},NotBefore:time.Now().Add(-time.Hour),NotAfter:time.Now().AddDate(1,0,0),KeyUsage:x509.KeyUsageDigitalSignature,ExtKeyUsage:[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},DNSNames:[]string{host,"localhost"}}
 addrs,_:=net.InterfaceAddrs()
 for _,a:=range addrs { ip,_,err:=net.ParseCIDR(a.String()); if err==nil { t.IPAddresses=append(t.IPAddresses,ip) } }
 der,err:=x509.CreateCertificate(rand.Reader,&t,&t,&k.PublicKey,k); must(err)
 kb,err:=x509.MarshalPKCS8PrivateKey(k); must(err)
 must(os.WriteFile(key,pem.EncodeToMemory(&pem.Block{Type:"PRIVATE KEY",Bytes:kb}),0600))
 must(os.WriteFile(cert,pem.EncodeToMemory(&pem.Block{Type:"CERTIFICATE",Bytes:der}),0600))
 return cert,key
}
// DSM settings are read, never changed. The browser launcher also uses location.host.
func dsmPorts(path string) map[string]int {
 result:=map[string]int{}
 b,err:=os.ReadFile(path); if err!=nil { return result }
 for _,line:=range strings.Split(string(b),"\n") {
  k,v,ok:=strings.Cut(strings.TrimSpace(line),"="); if !ok { continue }
  if k!="admin_port" && k!="secure_admin_port" { continue }
  n,e:=strconv.Atoi(strings.Trim(v,"\"' \r")); if e==nil && n>0 && n<65536 { result[k]=n }
 }
 return result
}
func runChild(ctx context.Context, binary string) {
 for ctx.Err()==nil {
  mu.Lock(); setupCode=""; setupAt=time.Time{}; childStatus="starting"; mu.Unlock()
  cmd:=exec.Command(binary,"--config",filepath.Join(state,"config","config.yaml"))
  cmd.Dir=state
  // Kill the process group too: extensions must not survive Package Center Stop.
  cmd.SysProcAttr=&syscall.SysProcAttr{Setpgid:true}
  pipe,err:=cmd.StdoutPipe(); if err!=nil { log.Print(err); return }; cmd.Stderr=cmd.Stdout
  if err=cmd.Start(); err!=nil { log.Print(err); mu.Lock(); childStatus="failed to start"; mu.Unlock(); select {case <-ctx.Done():return;case <-time.After(10*time.Second):continue} }
  mu.Lock(); childStatus="running"; mu.Unlock()
  finished:=make(chan struct{})
  go func(){ select {
   case <-ctx.Done():
    _=syscall.Kill(-cmd.Process.Pid,syscall.SIGTERM)
    select {case <-finished:case <-time.After(20*time.Second):_=syscall.Kill(-cmd.Process.Pid,syscall.SIGKILL)}
   case <-finished:
  }}()
  scanner:=bufio.NewScanner(pipe); scanner.Buffer(make([]byte,65536),1024*1024)
  for scanner.Scan() {
   line:=scanner.Text()
   if m:=codePattern.FindStringSubmatch(line); len(m)==2 { mu.Lock(); setupCode=m[1]; setupAt=time.Now(); mu.Unlock() }
   // Do not persist upstream output: it contains the one-time secret.
  }
  if err:=scanner.Err(); err!=nil { log.Printf("upstream output reader: %v",err) }
  err=cmd.Wait(); close(finished)
  _=syscall.Kill(-cmd.Process.Pid,syscall.SIGKILL)
  mu.Lock(); setupCode=""; childStatus="restarting"; mu.Unlock()
  if ctx.Err()!=nil { return }
  log.Printf("upstream exited (%v); retrying in 5 seconds",err)
  select {case <-ctx.Done():return;case <-time.After(5*time.Second):}
 }
}
var page=template.Must(template.New("setup").Parse(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Prusa Connect Local | Synology</title><style>body{font:16px system-ui;background:#f4f6fa;color:#17202e;max-width:780px;margin:6vh auto;padding:24px}main{background:white;padding:32px;border-radius:20px}code{font-size:1.35rem}a{color:#b8430c}small{color:#536174}</style><main><select id="language"><option value="en">English</option><option value="cs">Čeština</option></select><h1>Prusa Connect Local</h1><p data-en="Unofficial Synology integration" data-cs="Neoficiální integrace pro Synology">Unofficial Synology integration</p><h2 data-en="One-time setup code" data-cs="Jednorázový instalační kód">One-time setup code</h2><code>{{.Code}}</code><p data-en="If the code is missing, wait a few seconds and refresh. After setup it is no longer needed. Never share it." data-cs="Pokud kód chybí, chvíli počkejte a obnovte stránku. Po nastavení už není potřeba. Nikomu jej nesdělujte.">If the code is missing, wait a few seconds and refresh. After setup it is no longer needed. Never share it.</p><p><a href="/" data-en="Open Prusa Connect (HTTPS)" data-cs="Otevřít Prusa Connect (HTTPS)">Open Prusa Connect (HTTPS)</a></p><p>Status: {{.Status}}</p><small>DSM detected ports: {{.Ports}}<br>HTTPS: 18443 | Printer HTTP: 18080</small><p data-en="Use only on a trusted LAN. The HTTP printer listener is not encrypted. This package does not configure your firewall or router." data-cs="Používejte jen v důvěryhodné LAN. HTTP spojení tiskáren není šifrované. Balíček nemění firewall ani router.">Use only on a trusted LAN. The HTTP printer listener is not encrypted. This package does not configure your firewall or router.</p></main><script>const s=document.querySelector('#language');function lang(l){document.documentElement.lang=l;document.querySelectorAll('[data-en]').forEach(e=>e.textContent=e.dataset[l]);localStorage.setItem('pcl-language',l)}s.value=localStorage.getItem('pcl-language')||(navigator.language.startsWith('cs')?'cs':'en');s.onchange=()=>lang(s.value);lang(s.value)</script></html>`))
func main() {
 state=os.Getenv("SYNOPKG_PKGVAR"); if state=="" { state="./state" }; must(os.MkdirAll(state,0700)); state,_=filepath.Abs(state)
 if len(os.Args)>1 && os.Args[1]=="init" { initCredential(); return }
 if len(os.Args)<2 { log.Fatal("usage: pcl-wrapper /path/to/prusa_connect_amd64 | init") }
 must(os.MkdirAll(filepath.Join(state,"config"),0700))
 cfg:=filepath.Join(state,"config","config.yaml")
 if _,err:=os.Stat(cfg); os.IsNotExist(err) {
  // The browser listener is HTTPS; HTTP is retained separately for printer firmware.
  must(os.WriteFile(cfg,[]byte(fmt.Sprintf("core:\n  api:\n    addr: \":%d\"\n  trusted_proxies:\n    - 127.0.0.1/32\n    - ::1/128\n  upgrade:\n    enabled: false\n",backendPort)),0600))
 }
 cert,key:=certFiles()
 // Bind before launching upstream, so a conflicting HTTPS port fails cleanly.
 listener,err:=net.Listen("tcp",fmt.Sprintf(":%d",httpsPort)); must(err)
 ctx,cancel:=signal.NotifyContext(context.Background(),syscall.SIGTERM,syscall.SIGINT); defer cancel()
 done:=make(chan struct{}); go func(){ defer close(done); runChild(ctx,os.Args[1]) }()
 target,_:=url.Parse(fmt.Sprintf("http://127.0.0.1:%d",backendPort))
 proxy:=httputil.NewSingleHostReverseProxy(target)
 proxy.FlushInterval=-1
 original:=proxy.Director
 proxy.Director=func(r *http.Request){ original(r); r.Header.Del("Forwarded"); r.Header.Del("X-Forwarded-For"); r.Header.Set("X-Forwarded-Proto","https"); r.Header.Set("X-Forwarded-Host",r.Host) }
 proxy.ErrorHandler=func(w http.ResponseWriter,r *http.Request,e error){ http.Error(w,"Prusa Connect is starting or unavailable. Retry shortly.",503) }
 mux:=http.NewServeMux()
 mux.Handle("/_synology/setup",auth(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store"); w.Header().Set("Referrer-Policy","no-referrer"); w.Header().Set("X-Frame-Options","DENY")
  mu.RLock(); code,status,at:=setupCode,childStatus,setupAt; mu.RUnlock()
  if code=="" || time.Since(at)>30*time.Minute { code="Not available / Není dostupný" }
  // Setup codes are displayed for at most 30 minutes and only behind portal auth.
  _=page.Execute(w,struct{Code,Status string; Ports map[string]int}{code,status,dsmPorts("/etc/synoinfo.conf")})
 })))
 mux.Handle("/",proxy)
 srv:=&http.Server{Handler:mux,ReadHeaderTimeout:10*time.Second,IdleTimeout:90*time.Second,MaxHeaderBytes:1<<20,TLSConfig:&tls.Config{MinVersion:tls.VersionTLS12}}
 go func(){ <-ctx.Done(); c,stop:=context.WithTimeout(context.Background(),25*time.Second); defer stop(); _=srv.Shutdown(c) }()
 err=srv.ServeTLS(listener,cert,key)
 if err!=nil && err!=http.ErrServerClosed { log.Print(err); cancel() }
 cancel(); <-done
}
