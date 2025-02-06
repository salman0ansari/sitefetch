# sitefetch

Fetch a site and extract its readable content as Markdown (to be used with AI models).

![image](https://github.com/user-attachments/assets/bdd90bfe-ed4e-445f-8121-b8284128e0c4)

## Install

Install globally

```bash
go install github.com/salman0ansari/sitefetch@latest
```

## Usage

![image](https://github.com/user-attachments/assets/4391566d-bb82-4089-a3d8-0ae3b3db46ae)

```bash
sitefetch https://hisalman.in --outfile site.txt

# or better concurrency
sitefetch https://hisalman.in --outfile site.txt --concurrency 10
```

### Match specific pages

Use the `--match` flag to specify the pages you want to fetch:

```bash
sitefetch https://vite.dev -m "/blog/**" -m "/guide/**"
```

### Content selector

```bash
sitefetch https://vite.dev --content-selector ".content"
```

## Credit 
This project is a Go implementation of the original [sitefetch](https://github.com/egoist/sitefetch) written in TypeScript by [egoist](https://github.com/egoist).
