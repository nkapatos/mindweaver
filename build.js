const esbuild = require('esbuild');
const postcss = require('postcss');
const tailwindcss = require('@tailwindcss/postcss');
const fs = require('fs');

function postCSSPlugin() {
  return {
    name: 'postcss',
    setup(build) {
      build.onLoad({ filter: /\.css$/ }, async args => {
        const css = fs.readFileSync(args.path, 'utf8');
        const result = await postcss([tailwindcss]).process(css, { from: args.path });
        return { contents: result.css, loader: 'css' };
      });
    },
  };
}

async function build() {
  const ctx = await esbuild.context({
    entryPoints: {
      'main': 'assets/ts/main.ts',
      'index': 'assets/css/index.css',
    },
    bundle: true,
    outdir: 'dist',
    loader: { '.css': 'css' },
    sourcemap: true,
    plugins: [postCSSPlugin()],
  });

  await ctx.watch();
  console.log('Watching for changes...');
}

build().catch(err => {
  console.error(err);
  process.exit(1);
});