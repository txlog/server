/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './templates/**/*.html',
  ],
  safelist: [
    // Dynamic classes used in JS via string concatenation (e.g., 'bg-' + color)
    // analytics_anomalies.html and index.html build classes dynamically
    'bg-txlog-coral/5', 'bg-txlog-coral/10', 'border-txlog-coral/20', 'text-txlog-coral',
    'bg-txlog-golden/5', 'bg-txlog-golden/10', 'border-txlog-golden/20', 'text-txlog-golden',
    'bg-txlog-sky/5', 'bg-txlog-sky/10', 'border-txlog-sky/20', 'text-txlog-sky',
  ],
  theme: {
    extend: {
      colors: {
        txlog: {
          indigo: '#424565',
          lavender: '#E6E6FA',
          coral: '#D9556A',
          golden: '#F4B54B',
          sky: '#4A8AE8',
          leaf: '#4A9E42',
          purple: '#8B5CF6',
          bg: '#F8F9FE',
        }
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        display: ['Poppins', 'sans-serif'],
      },
      boxShadow: {
        'soft': '0 8px 24px rgba(66, 69, 101, 0.08)',
        'glow-sky': '0 0 0 4px rgba(106, 162, 251, 0.2)',
      },
      borderRadius: {
        'xl': '12px',
        '2xl': '16px',
        '3xl': '24px',
      }
    }
  },
  plugins: [],
}
