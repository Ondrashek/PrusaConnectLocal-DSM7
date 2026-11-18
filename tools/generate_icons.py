#!/usr/bin/env python3
"""Resize icon/icon.png for DSM. Run automatically by build_spk.py."""
import argparse
from pathlib import Path
from PIL import Image, ImageOps
SIZES = (16, 24, 32, 48, 64, 72, 256)
def generate(root, source=None):
    root = Path(root)
    source = Path(source) if source else root / 'icon' / 'icon.png'
    if not source.is_file():
        raise FileNotFoundError(f'Place your image at {source}')
    with Image.open(source) as image:
        image = ImageOps.exif_transpose(image).convert('RGBA')
        # Preserve aspect ratio; pad non-square input with transparency.
        side = max(image.size)
        square = Image.new('RGBA', (side, side), (0, 0, 0, 0))
        square.alpha_composite(image, ((side-image.width)//2, (side-image.height)//2))
        folder = root / 'spk' / 'ui' / 'images'
        folder.mkdir(parents=True, exist_ok=True)
        for size in SIZES:
            square.resize((size,size), Image.Resampling.LANCZOS).save(folder / f'app_{size}.png')
        for size, name in ((72,'PACKAGE_ICON.PNG'),(256,'PACKAGE_ICON_256.PNG')):
            square.resize((size,size), Image.Resampling.LANCZOS).save(root / 'spk' / name)
    return source
if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--source', type=Path)
    args = parser.parse_args()
    generate(Path(__file__).resolve().parents[1], args.source)
    print('DSM icons generated successfully.')
