"""Text-only g4f: block writing images/audio to disk."""

from __future__ import annotations

import g4f.image.copy_images as copy_images


async def _blocked_save_response_media(*_args, **_kwargs):
    if False:
        yield


async def _passthrough_copy_media(images, *_args, **_kwargs):
    return list(images)


def disable_g4f_media_writes() -> None:
    copy_images.ensure_media_dir = lambda: None
    copy_images.save_response_media = _blocked_save_response_media
    copy_images.copy_media = _passthrough_copy_media
