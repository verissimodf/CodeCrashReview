/** @type {import('tailwindcss').Config} */
export default {
	content: [
		'./src/**/*.{html,js,svelte,ts}',
		'./node_modules/lucide-svelte/**/*.js'
	],
	darkMode: 'class',
	theme: {
		extend: {
			colors: {
				// Cyberpunk/Neon palette
				cyber: {
					bg: '#0d1117',
					bgSecondary: '#161b22',
					bgTerritary: '#21262d',
					border: '#30363d',
					borderBright: '#484f58',
				},
				neon: {
					cyan: '#00ffff',
					cyanGlow: '#00ffff40',
					magenta: '#ff00ff',
					magentaGlow: '#ff00ff40',
					green: '#00ff87',
					greenGlow: '#00ff8740',
					amber: '#ffaa00',
					amberGlow: '#ffaa0040',
					red: '#ff3366',
					redGlow: '#ff336640',
					blue: '#0099ff',
					blueGlow: '#0099ff40',
				},
				agent: {
					sentinel: '#ff3366',
					optimizer: '#00ff87',
					refactor: '#0099ff',
					guardian: '#ffaa00',
					docubot: '#ff00ff',
					polyglot: '#00ffff',
				},
				rating: {
					unranked: '#8b949e',
					bronze: '#cd7f32',
					silver: '#c0c0c0',
					gold: '#ffd700',
					platinum: '#e5e4e2',
					diamond: '#b9f2ff',
					master: '#ff6b6b',
					grandmaster: '#ff3366',
				}
			},
			fontFamily: {
				mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'monospace'],
				display: ['Orbitron', 'Space Grotesk', 'system-ui', 'sans-serif'],
				ui: ['IBM Plex Sans', 'Inter', 'system-ui', 'sans-serif'],
			},
			fontSize: {
				'xs': ['0.7rem', { lineHeight: '1.4' }],
				'sm': ['0.8125rem', { lineHeight: '1.5' }],
				'base': ['0.9375rem', { lineHeight: '1.6' }],
				'lg': ['1.0625rem', { lineHeight: '1.6' }],
				'xl': ['1.25rem', { lineHeight: '1.5' }],
				'2xl': ['1.5rem', { lineHeight: '1.4' }],
				'3xl': ['2rem', { lineHeight: '1.3' }],
				'4xl': ['2.5rem', { lineHeight: '1.2' }],
				'5xl': ['3.5rem', { lineHeight: '1.1' }],
			},
			spacing: {
				'18': '4.5rem',
				'88': '22rem',
				'128': '32rem',
			},
			animation: {
				'glow': 'glow 2s ease-in-out infinite alternate',
				'pulse-glow': 'pulse-glow 3s ease-in-out infinite',
				'scanline': 'scanline 8s linear infinite',
				'grid-move': 'grid-move 20s linear infinite',
				'float': 'float 6s ease-in-out infinite',
				'slide-up': 'slide-up 0.3s ease-out',
				'slide-down': 'slide-down 0.3s ease-out',
				'fade-in': 'fade-in 0.2s ease-out',
				'scale-in': 'scale-in 0.2s ease-out',
				'shake': 'shake 0.5s ease-in-out',
				'typewriter': 'typewriter 2s steps(40) 1s forwards',
				'blink': 'blink 1s step-end infinite',
			},
			keyframes: {
				glow: {
					'0%': { textShadow: '0 0 5px currentColor, 0 0 10px currentColor, 0 0 15px currentColor' },
					'100%': { textShadow: '0 0 10px currentColor, 0 0 20px currentColor, 0 0 30px currentColor' },
				},
				'pulse-glow': {
					'0%, 100%': { boxShadow: '0 0 5px currentColor, 0 0 10px currentColor' },
					'50%': { boxShadow: '0 0 20px currentColor, 0 0 30px currentColor' },
				},
				scanline: {
					'0%': { transform: 'translateY(-100%)' },
					'100%': { transform: 'translateY(100%)' },
				},
				'grid-move': {
					'0%': { backgroundPosition: '0 0' },
					'100%': { backgroundPosition: '100px 100px' },
				},
				float: {
					'0%, 100%': { transform: 'translateY(0)' },
					'50%': { transform: 'translateY(-10px)' },
				},
				'slide-up': {
					'0%': { opacity: '0', transform: 'translateY(10px)' },
					'100%': { opacity: '1', transform: 'translateY(0)' },
				},
				'slide-down': {
					'0%': { opacity: '0', transform: 'translateY(-10px)' },
					'100%': { opacity: '1', transform: 'translateY(0)' },
				},
				'fade-in': {
					'0%': { opacity: '0' },
					'100%': { opacity: '1' },
				},
				'scale-in': {
					'0%': { opacity: '0', transform: 'scale(0.95)' },
					'100%': { opacity: '1', transform: 'scale(1)' },
				},
				shake: {
					'0%, 100%': { transform: 'translateX(0)' },
					'25%': { transform: 'translateX(-5px)' },
					'75%': { transform: 'translateX(5px)' },
				},
				typewriter: {
					'0%': { width: '0' },
					'100%': { width: '100%' },
				},
				blink: {
					'0%, 50%': { opacity: '1' },
					'51%, 100%': { opacity: '0' },
				},
			},
			boxShadow: {
				'neon-cyan': '0 0 5px #00ffff, 0 0 10px #00ffff, 0 0 20px #00ffff',
				'neon-magenta': '0 0 5px #ff00ff, 0 0 10px #ff00ff, 0 0 20px #ff00ff',
				'neon-green': '0 0 5px #00ff87, 0 0 10px #00ff87, 0 0 20px #00ff87',
				'neon-amber': '0 0 5px #ffaa00, 0 0 10px #ffaa00, 0 0 20px #ffaa00',
				'neon-red': '0 0 5px #ff3366, 0 0 10px #ff3366, 0 0 20px #ff3366',
				'glow-sm': '0 0 5px currentColor',
				'glow-md': '0 0 10px currentColor, 0 0 20px currentColor',
				'glow-lg': '0 0 20px currentColor, 0 0 40px currentColor',
				'card': '0 4px 6px -1px rgb(0 0 0 / 0.3), 0 2px 4px -2px rgb(0 0 0 / 0.2)',
				'card-hover': '0 10px 15px -3px rgb(0 0 0 / 0.4), 0 4px 6px -4px rgb(0 0 0 / 0.3)',
			},
			backgroundImage: {
				'grid-pattern': 'linear-gradient(rgba(0, 255, 255, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 255, 255, 0.03) 1px, transparent 1px)',
				'radial-glow': 'radial-gradient(ellipse at center, var(--tw-gradient-from) 0%, transparent 70%)',
				'scanlines': 'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0, 0, 0, 0.1) 2px, rgba(0, 0, 0, 0.1) 4px)',
			},
			backgroundSize: {
				'grid': '50px 50px',
			},
			transitionDuration: {
				'fast': '150ms',
				'normal': '200ms',
				'slow': '300ms',
			},
			transitionTimingFunction: {
				'bounce-in': 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
				'smooth': 'cubic-bezier(0.4, 0, 0.2, 1)',
			},
			zIndex: {
				'60': '60',
				'70': '70',
				'80': '80',
				'90': '90',
				'100': '100',
			},
		},
	},
	plugins: [
		require('@tailwindcss/forms'),
		require('@tailwindcss/typography'),
		function({ addUtilities }) {
			const newUtilities = {
				'.text-glow': {
					textShadow: '0 0 5px currentColor, 0 0 10px currentColor',
				},
				'.text-glow-lg': {
					textShadow: '0 0 10px currentColor, 0 0 20px currentColor, 0 0 30px currentColor',
				},
				'.border-glow': {
					boxShadow: 'inset 0 0 10px currentColor, 0 0 10px currentColor',
				},
				'.scanlines': {
					position: 'relative',
					'&::before': {
						content: '""',
						position: 'absolute',
						inset: '0',
						backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0, 0, 0, 0.1) 2px, rgba(0, 0, 0, 0.1) 4px)',
						pointerEvents: 'none',
						zIndex: '10',
					},
				},
				'.glass': {
					background: 'rgba(22, 27, 34, 0.8)',
					backdropFilter: 'blur(12px)',
					border: '1px solid rgba(48, 54, 61, 0.5)',
				},
				'.glass-strong': {
					background: 'rgba(13, 17, 23, 0.9)',
					backdropFilter: 'blur(20px)',
					border: '1px solid rgba(72, 79, 88, 0.3)',
				},
			};
			addUtilities(newUtilities);
		}
	],
};