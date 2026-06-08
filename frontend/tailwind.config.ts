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
      },
      keyframes: {
        'pulse-icon': {
          '0%, 100%': { opacity: '1', transform: 'scale(1)' },
          '50%': { opacity: '0.7', transform: 'scale(1.1)' },
        },
        'heartbeat': {
          '0%, 100%': { backgroundColor: 'rgb(220 252 231)' },
          '50%': { backgroundColor: 'rgb(187 247 208)' },
        },
        'heartbeat-dark': {
          '0%, 100%': { backgroundColor: 'rgb(20 83 45 / 0.3)' },
          '50%': { backgroundColor: 'rgb(22 101 52 / 0.4)' },
        },
      },
      animation: {
        'pulse-icon': 'pulse-icon 2s ease-in-out infinite',
        'heartbeat': 'heartbeat 1.5s ease-in-out infinite',
        'heartbeat-dark': 'heartbeat-dark 1.5s ease-in-out infinite',
      },
    }
  }
}
