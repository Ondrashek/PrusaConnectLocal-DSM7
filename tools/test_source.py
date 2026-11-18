"""Local checks, not a DSM runtime test."""
import json
from pathlib import Path
from PIL import Image
from generate_icons import generate, SIZES
root = Path(__file__).resolve().parents[1]
generate(root)
for size in SIZES:
    with Image.open(root / 'spk/ui/images' / f'app_{size}.png') as image:
        assert image.size == (size,size)
for size,name in ((72,'PACKAGE_ICON.PNG'),(256,'PACKAGE_ICON_256.PNG')):
    with Image.open(root / 'spk' / name) as image:
        assert image.size == (size,size)
for path in (root/'spk/WIZARD_UIFILES').glob('install_uifile*'):
    data=json.loads(path.read_text())
    assert data[0]['items'][0]['subitems'][0]['validator']['minLength']==8
assert '-ge 8' in (root/'spk/scripts/preinst').read_text()
assert 'len(p)<8' in (root/'cmd/pcl-wrapper/main.go').read_text()
launcher=(root/'spk/ui/index.html').read_text()
assert "target.port = '18443'" in launcher
assert 'window.location.replace(target.href)' in launcher
print('Source checks and icon generation passed; no Go or DSM runtime test performed.')
