"""Source and icon checks; not a DSM runtime test."""
import json
from pathlib import Path
from PIL import Image
from generate_icons import generate, SIZES
root=Path(__file__).resolve().parents[1]
generate(root)
for size in SIZES:
    with Image.open(root/'spk/ui/images'/f'app_{size}.png') as image:
        assert image.size==(size,size)
for p in (root/'spk/WIZARD_UIFILES').glob('install_uifile*'):
    data=json.loads(p.read_text())
    assert data[0]['items'][0]['subitems'][0]['validator']['minLength']==8
main=(root/'cmd/pcl-wrapper/main.go').read_text()
assert 'len(p)<8' in main
assert 'httputil' not in main
assert 'const portalPort = 18444' in main
launcher=(root/'spk/ui/index.html').read_text()
assert "u.port='18443'" in launcher
assert 'location.replace(u.href)' in launcher
print('Source checks passed; run Go tests and DSM tests separately.')
