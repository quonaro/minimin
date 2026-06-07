import type { Config } from 'tailwindcss'

export default <Partial<Config>>{
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#0052FF',
          50: '#e6f0ff',
          100: '#b3d1ff',
          200: '#80b3ff',
          300: '#4d94ff',
          400: '#1a75ff',
          500: '#0052FF',
          600: '#0042cc',
          700: '#003199',
          800: '#002166',
          900: '#001033',
        },
        background: {
          dark: '#0F0F10',
        },
        border: {
          dark: 'rgba(255, 255, 255, 0.15)',
        }
      }
    }
  }
}
