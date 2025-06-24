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

function onEndPlugin(host, port) {
  return {
    name: 'onEnd',
    setup(build) {
      build.onEnd(() => {
        console.log('esbuild completed successfully');
        const { exec } = require('child_process');
        exec(`templ generate --notify-proxy --proxybind="${host}" --proxyport="${port}"`, (error, stdout, stderr) => {
          if (error) {
            console.error('Error executing templ command:', error);
            return;
          }
          if (stderr) {
            console.error('Templ command stderr:', stderr);
            return;
          }
          console.log('notified templ to reload');
        });
      });
    },
  };
}

function getBuildConfig(host, port) {
  return {
    entryPoints: {
      'main': 'assets/ts/main.ts',
      'index': 'assets/css/index.css',
    },
    bundle: true,
    outdir: 'dist',
    loader: { '.css': 'css' },
    sourcemap: true,
    plugins: [postCSSPlugin(), onEndPlugin(host, port)],
  };
}

async function dev() {
  const host = process.env.APP_HOST || 'localhost';
    const port = process.env.PROXY_PORT || '8081';
  const config = getBuildConfig(host, port);
  const ctx = await esbuild.context(config);
  
  await ctx.watch({});
  console.log('Watching for changes...');
  console.log(`Templ proxy notifications will be sent to ${host}:${port}`);
}

async function build() {
  const config = getBuildConfig();
  
  try {
    const result = await esbuild.build(config);
    console.log('Build completed successfully');
    return result;
  } catch (error) {
    console.error('Build failed:', error);
    throw error;
  }
}

// CLI handling
const command = process.argv[2];

if (command === 'dev') {
  dev().catch(err => {
    console.error(err);
    process.exit(1);
  });
} else if (command === 'build') {
  build().catch(err => {
    console.error(err);
    process.exit(1);
  });
} else {
  console.log('Usage: node build.js [dev|build]');
  console.log('  dev   - Start development mode with file watching');
  console.log('  build - Build for production');
  process.exit(1);
}
