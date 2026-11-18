import hashlib, io, json, sys, tarfile
with tarfile.open(sys.argv[1]) as t:
    names=t.getnames()
    for n in ['INFO','package.tgz','conf/privilege','WIZARD_UIFILES/install_uifile','scripts/start-stop-status']:
        assert n in names,n
    for n in names:
        assert not n.startswith('/') and '..' not in n.split('/'),n
        if n.startswith(('conf/','WIZARD_UIFILES/')):json.load(t.extractfile(n))
    info=t.extractfile('INFO').read().decode()
    payload=t.extractfile('package.tgz').read()
    assert 'checksum="'+hashlib.md5(payload).hexdigest()+'"' in info
    with tarfile.open(fileobj=io.BytesIO(payload),mode='r:gz') as p:
        for n in ['bin/pcl-wrapper','bin/prusa_connect_amd64','ui/config','ui/index.html']:
            assert n in p.getnames(),n
        json.load(p.extractfile('ui/config'))
print('Archive structure, JSON, and payload checksum OK. This is NOT a DSM installation test.')
