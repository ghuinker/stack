import os
import json
import requests

from django import template
from django.utils.html import format_html
from django.utils.safestring import mark_safe
from django.conf import settings


register = template.Library()


def _get_manifest():
    file_name = os.path.join(settings.DIST_DIR, "static/assets/manifest.json")
    with open(file_name, "r") as f:
        return json.load(f)


@register.simple_tag
def asset_manifest(name: str) -> str:
    if settings.DEBUG:
        return "http://localhost:3000/static/" + name
    entry = manifest.get(name)
    if entry and (file := entry.get("file")):
        return "/static/assets/" + file
    print(f"Asset manifest not found for: {name}")
    return ""


manifest = _get_manifest()
