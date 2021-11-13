const { loadConfigFromFile, mergeConfig } = require("vite")
const path = require("path");

module.exports = {
  "stories": [
    "../src/**/*.stories.mdx",
    "../src/**/*.stories.@(js|jsx|ts|tsx)"
  ],
  "addons": [
    "@storybook/addon-actions",
    "@storybook/addon-docs",
    "@storybook/addon-controls",
    "@storybook/addon-essentials",
    "@storybook/addon-links",
    "@storybook/addon-viewport",
    // "@storybook/source-loader",
  ],
  "core": {
    "builder": "storybook-builder-vite"
  },
  
  // customize the vite config for storybook here
  // See also: https://github.com/eirslett/storybook-builder-vite
  // configType is "DEVELOPMENT" or "PRODUCTION"
  // config is the builder's own configuration
  // returns the updated configuration
  async viteFinal(config, { configType }) {
  
    // Load vite's own configuration
    const { config: userConfig } = await loadConfigFromFile(
        path.resolve(__dirname, "../vite.config.ts")
    );

    // and merge it with the builder's configuration
    return mergeConfig(config, {
      ...userConfig,

      // Use our standard cache dir, to make it easy to clear
      cacheDir: path.resolve(__dirname, '../node_modules/.cache/vite'),

      // manually specify plugins to avoid conflict
      plugins: [
        '@storybook/addon-actions',
        "@storybook/addon-essentials",
        "@storybook/addon-knobs",
        "@storybook/addon-links"
      ],
      // optimizeDeps: {
      //   ...config.optimizeDeps,
      //   // Entries are specified relative to the root
      //   entries: [`${path.relative(config.root, path.resolve(__dirname, '../src'))}/**/*.stories.ts`],
      //   include: [...config.optimizeDeps.include, 'storybook-dark-mode'],
      // },
    });
  },
}