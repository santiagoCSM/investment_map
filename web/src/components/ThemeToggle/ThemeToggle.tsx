import { useTheme } from '../../hooks/useTheme'

export function ThemeToggle() {
  const { theme, toggle } = useTheme()

  return (
    <button
      type="button"
      className="fixed top-5 left-5 z-20 flex items-center justify-center size-[34px] p-0 border border-[var(--panel-border)] rounded-[10px] bg-[var(--panel-bg)] backdrop-blur-[14px] text-[var(--panel-text)] cursor-pointer transition-all duration-150 hover:scale-105 active:scale-95"
      onClick={toggle}
      aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
      title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      {theme === 'dark' ? (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
        </svg>
      ) : (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
          <path d="M12.3 2a.9.9 0 0 1 .9 1.5A8 8 0 0 0 20.5 11a.9.9 0 0 1 1.5.9A10 10 0 1 1 12.3 2Z" />
        </svg>
      )}
    </button>
  )
}
