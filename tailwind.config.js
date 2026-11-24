/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}", "./frontend/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      // ===== COLORS =====
      colors: {
        // Primary Brand
        primary: {
          DEFAULT: '#3b82f6',
          light: '#60a5fa',
          dark: '#2563eb',
          deepest: '#1d4ed8',
          darkest: '#1e40af',
        },
        // Navy Gradient Family (Dark Mode / Premium Content)
        navy: {
          dark: '#0f0f23',
          medium: '#1a1a2e',
          light: '#16213e',
          overlay: 'rgba(30, 30, 50, 0.8)',
          border: '#2a2a40',
          deep: '#1a1a2e',
          darkest: '#0a0a18',
          brownDark: '#0f0f23',
        },
        // Semantic Colors
        success: {
          DEFAULT: '#10b981',
          light: '#34d399',
          dark: '#065f46',
          darker: '#38a169',
          bgLight: '#d1fae5',
        },
        danger: {
          brightest: '#ff4444',
          brightest: '#f87171',
          bright: '#ef4444',
          DEFAULT: '#dc2626',
          dark: '#b91c1b',
          darker: '#991b1b',
        },
        warning: {
          brightest: '#fcd34a',
          bright: '#fbbf24',
          DEFAULT: '#f59e0b',
          dark: '#d97706',
          bgLight: '#fffbeb',
        },
        // Victory / Gold
        victory: {
          gold: '#ffd700',
          light: '#ffed4e',
        },
        // Gray Scale
        gray: {
          50: '#f9fafb',
          100: '#f5f7fa',
          200: '#f5f5f5',
          300: '#d1d5db',
          400: '#9ca3af',
          500: '#6b7280',
          600: '#718096',
          700: '#4a5568',
          800: '#2d3748',
        },
        white: '#ffffff',
        black: '#000000',
        // Combat Stats
        strength: {
          DEFAULT: '#ef4444',
          dark: '#dc2626',
        },
        defense: {
          DEFAULT: '#3b82f6',
          dark: '#2563eb',
        },
        agility: {
          DEFAULT: '#eab308',
          dark: '#ca8a04',
        },
        health: '#10b981',
        energy: '#fbbf24',
        // Level Badges
        level: {
          1: { start: '#718096', end: '#4a5568' },
          2: { start: '#38b2ac', end: '#319795' },
          3: { start: '#dd6b20', end: '#c05621' },
          4: { start: '#805ad5', end: '#704ada' },
          5: { start: '#d69e2e', end: '#fbbf24' },
        },
        // Accents
        accent: {
          purple: '#7c3aed',
          purpleDark: '#6d28d9',
          purpleLight: '#a78bfa',
          indigo: '#6366f1',
          teal: '#14b8a6',
        },
        // Arena / Theater Mode
        arena: {
          dark: '#4a2a2a',
          darker: '#2d1a1a',
          charcoal: '#151520',
        },
        // Slate
        slate: {
          dark: '#1e293b',
          darker: '#334155',
        },
        // Text
        text: {
          muted: '#d1d5db',
          bright: '#e5e7eb',
        },
      },

      // ===== SCREENS =====
      screens: {
        'sm': '480px',
        'md': '768px',
        'lg': '1024px',
        'xl': '1440px',
      },

      // ===== SPACING =====
      spacing: {
        '1': '0.25rem',   // 4px
        '2': '0.5rem',    // 8px
        '3': '0.75rem',   // 12px
        '4': '1rem',      // 16px
        '5': '1.25rem',   // 20px
        '6': '1.5rem',    // 24px
        '8': '2rem',      // 32px
        '10': '2.5rem',   // 40px
        '12': '3rem',     // 48px
      },

      // ===== BORDER RADIUS =====
      borderRadius: {
        sm: '6px',
        md: '8px',
        lg: '10px',
        xl: '12px',
        '2xl': '16px',
        '3xl': '20px',
        full: '9999px',
      },

      // ===== FONT FAMILY =====
      fontFamily: {
        sans: ['Segoe UI', 'Tahoma', 'Geneva', 'Verdana', 'sans-serif'],
      },

      // ===== FONT SIZES =====
      fontSize: {
        xs: '0.65rem',
        sm: '0.875rem',
        base: '1rem',
        lg: '1.25rem',
        xl: '1.5rem',
        'xxl': '2rem',
      },

      // ===== FONT WEIGHTS =====
      fontWeight: {
        normal: '400',
        medium: '500',
        semibold: '600',
        bold: '700',
        extrabold: '800',
      },

      // ===== LINE HEIGHTS =====
      lineHeight: {
        tight: '1.2',
        normal: '1.6',
        relaxed: '1.8',
      },

      // ===== SHADOWS =====
      boxShadow: {
        subtle: '0 1px 3px rgba(0, 0, 0, 0.1)',
        card: '0 4px 20px rgba(0, 0, 0, 0.08)',
        elevated: '0 8px 30px rgba(0, 0, 0, 0.12)',
        topbar: '0 2px 20px rgba(0, 0, 0, 0.3)',
        // Glow effects
        'glow-red': '0 0 10px rgba(239, 68, 68, 0.5)',
        'glow-blue': '0 0 10px rgba(59, 130, 246, 0.5)',
        'glow-gold': '0 0 10px rgba(234, 179, 8, 0.5)',
      },

      // ===== TRANSITIONS =====
      transitionDuration: {
        fast: '150ms',
        normal: '250ms',
        slow: '400ms',
      },
      transitionTimingFunction: {
        default: 'cubic-bezier(0.4, 0, 0.2, 1)',
        bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
      },

      // ===== Z-INDEX =====
      zIndex: {
        base: '0',
        dropdown: '100',
        sticky: '200',
        overlay: '300',
        modal: '400',
        popover: '500',
        tooltip: '600',
      },

      // ===== CUSTOM UTILITIES =====
      topbar: {
        height: '64px',
      },
    },
  },
  plugins: [],
}
