/** @type {import('tailwindcss').Config} */
export default {
    darkMode: ["class"],
    content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
  	extend: {
  		colors: {
  			primary: {
  				'50': '#f0f9ff',
  				'100': '#e0f2fe',
  				'200': '#bae6fd',
  				'300': '#7dd3fc',
  				'400': '#38bdf8',
  				'500': '#0ea5e9',
  				'600': '#0284c7',
  				'700': '#0369a1',
  				'800': '#075985',
  				'900': '#0c4a6e',
  				DEFAULT: 'hsl(var(--primary))',
  				foreground: 'hsl(var(--primary-foreground))'
  			},
  			secondary: {
  				'50': '#f7fee7',
  				'100': '#ecfccb',
  				'200': '#d9f99d',
  				'300': '#bef264',
  				'400': '#a3e635',
  				'500': '#84cc16',
  				'600': '#65a30d',
  				'700': '#4d7c0f',
  				'800': '#365314',
  				'900': '#1a2e05',
  				DEFAULT: 'hsl(var(--secondary))',
  				foreground: 'hsl(var(--secondary-foreground))'
  			},
  			background: 'hsl(var(--background))',
  			foreground: 'hsl(var(--foreground))',
  			card: {
  				DEFAULT: 'hsl(var(--card))',
  				foreground: 'hsl(var(--card-foreground))'
  			},
  			popover: {
  				DEFAULT: 'hsl(var(--popover))',
  				foreground: 'hsl(var(--popover-foreground))'
  			},
  			muted: {
  				DEFAULT: 'hsl(var(--muted))',
  				foreground: 'hsl(var(--muted-foreground))'
  			},
  			accent: {
  				DEFAULT: 'hsl(var(--accent))',
  				foreground: 'hsl(var(--accent-foreground))'
  			},
  			destructive: {
  				DEFAULT: 'hsl(var(--destructive))',
  				foreground: 'hsl(var(--destructive-foreground))'
  			},
  			border: 'hsl(var(--border))',
  			input: 'hsl(var(--input))',
  			ring: 'hsl(var(--ring))',
  			chart: {
  				'1': 'hsl(var(--chart-1))',
  				'2': 'hsl(var(--chart-2))',
  				'3': 'hsl(var(--chart-3))',
  				'4': 'hsl(var(--chart-4))',
  				'5': 'hsl(var(--chart-5))'
  			}
  		},
  		fontFamily: {
  			sans: [
  				'Inter',
  				'ui-sans-serif',
  				'system-ui',
  				'sans-serif'
  			]
  		},
  				borderRadius: {
			lg: 'var(--radius)',
			md: 'calc(var(--radius) - 2px)',
			sm: 'calc(var(--radius) - 4px)'
		},
		animation: {
			first: "moveVertical 30s ease infinite",
			second: "moveInCircle 20s reverse infinite",
			third: "moveInCircle 40s linear infinite",
			fourth: "moveHorizontal 40s ease infinite",
			fifth: "moveInCircle 20s ease infinite",
		},
		keyframes: {
			moveHorizontal: {
				"0%": {
					transform: "translateX(-50%) translateY(-10%)",
				},
				"50%": {
					transform: "translateX(50%) translateY(10%)",
				},
				"100%": {
					transform: "translateX(-50%) translateY(-10%)",
				},
			},
			moveInCircle: {
				"0%": {
					transform: "rotate(0deg)",
				},
				"50%": {
					transform: "rotate(180deg)",
				},
				"100%": {
					transform: "rotate(360deg)",
				},
			},
			moveVertical: {
				"0%": {
					transform: "translateY(-50%)",
				},
				"50%": {
					transform: "translateY(50%)",
				},
				"100%": {
					transform: "translateY(-50%)",
				},
			},
		}
  	}
  },
  plugins: [require("tailwindcss-animate")],
} 