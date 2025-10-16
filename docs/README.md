# echoip Documentation

This directory contains the documentation for **echoip**, an IP address lookup service.

## Documentation Files

- **index.md** - Documentation homepage
- **API.md** - Complete API reference
- **SERVER.md** - Server administration guide
- **mkdocs.yml** - MkDocs configuration
- **requirements.txt** - Python dependencies for ReadTheDocs

## Building Documentation Locally

### Requirements

- Python 3.11+
- pip

### Setup

```bash
cd docs
pip install -r requirements.txt
```

### Preview

```bash
mkdocs serve
# Open http://localhost:8000
```

### Build

```bash
mkdocs build
# Output: site/
```

## Theme

This documentation uses the **Material for MkDocs** theme with the **Dracula** color scheme for a modern, dark-themed reading experience.

## Hosting

Documentation is automatically built and hosted on [Read the Docs](https://readthedocs.org) when pushed to the repository.

## Contributing

To improve documentation:

1. Edit the Markdown files in `docs/`
2. Preview changes locally with `mkdocs serve`
3. Submit a pull request

## License

MIT License - See ../LICENSE.md
