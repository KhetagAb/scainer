import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: '../scainer/api/openapi.yaml',
  output: 'src/client',
  plugins: ['@hey-api/client-fetch', '@tanstack/react-query'],
});
