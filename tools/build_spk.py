#!/usr/bin/env python3
"""Build an experimental unsigned DSM7 SPK. Requires two already-built binaries."""
import argparse, hashlib, io, json, tarfile
from generate_icons import generate
from pathlib import Path

def add(tar,name,data,mode=0o644):
    item=tarfile.TarInfo(name);item.size=len(data);item.mode=mode;item.mtime=0;item.uid=item.gid=0
    tar.addfile(item,io.BytesIO(data))

def build(wrapper, upstream, output):
    root=Path(__file__).resolve().parents[1]
    generate(root)
    for p in (wrapper,upstream):
        if p.read_bytes()[:4]!=b'\x7fELF': raise ValueError(f'Not an ELF executable: {p}')
    upstream_license=upstream.parent/'LICENSE'
    if not upstream_license.is_file() or upstream_license.stat().st_size==0:
        raise ValueError('Place the pinned upstream LICENSE beside the upstream binary')
    payload=io.BytesIO()
    with tarfile.open(fileobj=payload,mode='w:gz',format=tarfile.USTAR_FORMAT) as t:
        add(t,'LICENSES/integration-MIT.txt',(root/'LICENSE').read_bytes())
        add(t,'LICENSES/upstream-LICENSE.txt',upstream_license.read_bytes())
        add(t,'LICENSES/THIRD_PARTY.md',(root/'THIRD_PARTY.md').read_bytes())
        add(t,'bin/pcl-wrapper',wrapper.read_bytes(),0o755)
        add(t,'bin/prusa_connect_amd64',upstream.read_bytes(),0o755)
        for p in sorted((root/'spk/ui').rglob('*')):
            if p.is_file():add(t,'ui/'+p.relative_to(root/'spk/ui').as_posix(),p.read_bytes())

    desc='Experimental unofficial package. Open the package launcher for HTTPS and the protected setup code. Setup-code URL: https://<NAS-address>:18443/_synology/setup (admin + installation password). The launcher automatically uses the current DSM host and port. Self-signed TLS; LAN only. Printer HTTP: 18080.'
    info={'package':'PrusaConnectLocal','displayname':'Prusa Connect Local (unofficial beta)','version':'0.1.0-0004','arch':'geminilake','os_min_ver':'7.0-40000','maintainer':'Community integration','description':desc,'description_csy':'Experimentální neoficiální balíček. URL instalačního kódu: https://<adresa-NAS>:18443/_synology/setup (admin a heslo z instalace). Spouštěč přebírá adresu a port DSM automaticky. HTTPS s vlastním certifikátem. Jen důvěryhodná LAN. HTTP tiskáren: 18080.','thirdparty':'yes','beta':'yes','startable':'yes','dsmuidir':'ui','adminprotocol':'https','adminport':'18443','adminurl':'','silent_install':'no','silent_upgrade':'yes','silent_uninstall':'no','checksum':hashlib.md5(payload.getvalue()).hexdigest()}
    # INFO values are quoted; JSON escaping is compatible for these string values.
    raw='\n'.join(k+'='+json.dumps(v,ensure_ascii=False) for k,v in info.items())+'\n'
    output.parent.mkdir(parents=True,exist_ok=True)
    with tarfile.open(output,'w',format=tarfile.USTAR_FORMAT) as t:
        add(t,'INFO',raw.encode());add(t,'package.tgz',payload.getvalue())
        add(t,'PACKAGE_ICON.PNG',(root/'spk/PACKAGE_ICON.PNG').read_bytes());add(t,'PACKAGE_ICON_256.PNG',(root/'spk/PACKAGE_ICON_256.PNG').read_bytes())
        for folder in ('scripts','conf','WIZARD_UIFILES'):
            for p in sorted((root/'spk'/folder).rglob('*')):
                if p.is_file():add(t,folder+'/'+p.relative_to(root/'spk'/folder).as_posix(),p.read_bytes(),0o755 if folder=='scripts' else 0o644)
    output.with_suffix(output.suffix+'.sha256').write_text(hashlib.sha256(output.read_bytes()).hexdigest()+'  '+output.name+'\n')

if __name__=='__main__':
    a=argparse.ArgumentParser();a.add_argument('--wrapper',type=Path,required=True);a.add_argument('--upstream',type=Path,required=True);a.add_argument('--output',type=Path,required=True);v=a.parse_args();build(v.wrapper,v.upstream,v.output)
