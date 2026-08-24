/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        ink: '#05080c',
        panel: '#0c1218',
        line: '#1c2a33',
        phosphor: '#39ff88',
        amber: '#f0a202',
        ice: '#7ee8fa',
        danger: '#ff4d6a',
        mist: '#9bb0b8',
      },
      fontFamily: {
        display: ['"Chakra Petch"', 'sans-serif'],
        mono: ['"IBM Plex Mono"', 'ui-monospace', 'monospace'],
      },
    },
  },
  plugins: [],
}
