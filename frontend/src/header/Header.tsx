import { Export } from '@/components/export/Export'
import { Logo } from './Logo'
import type { AuthUser } from '@/api/auth'
import { UserMenu } from './UserMenu'

type HeaderProps = {
    graphId?: string;
    regionName?: string;
    user: AuthUser | null;
    onLogin: () => void;
    onLogout: () => void;
};

export function Header({ graphId, regionName, user, onLogin, onLogout }: HeaderProps) {
    return (
        <div className="grid w-full grid-cols-[1fr_auto_1fr] items-center gap-4">
            <Logo />
            <h1 className="max-w-[40vw] truncate text-center text-base font-semibold" title={regionName}>
                {regionName}
            </h1>
            <div className='flex justify-end gap-2'>
                <Export graphId={graphId} />
                <UserMenu user={user} onLogin={onLogin} onLogout={onLogout} />
            </div>
        </div>
    )
}
