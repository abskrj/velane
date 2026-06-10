import velaneLogo from '../assets/velane_sh.png'

type AuthBrandHeaderProps = {
  variant?: 'dark' | 'light'
  className?: string
}

export default function AuthBrandHeader({ variant = 'dark', className = '' }: AuthBrandHeaderProps) {
  return (
    <div className={`flex items-center gap-3 ${className}`}>
      <img
        src={velaneLogo}
        alt="Velane"
        className="h-10 w-10 shrink-0 rounded-xl object-contain"
      />
      <span
        className={`text-xl font-semibold tracking-tight ${variant === 'light' ? 'text-gray-900' : 'text-white'}`}
      >
        Velane
      </span>
    </div>
  )
}
