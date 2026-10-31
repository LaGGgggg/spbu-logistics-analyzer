import { Export } from '@/components/export/Export'
import { Logo } from './Logo'
import { Button } from '@/components/ui/button'
import { User } from 'lucide-react'

type HeaderProps = {
    graphId?: string;
    regionName?: string;
};

export function Header({ graphId, regionName }: HeaderProps) {
    return (
        <div className="grid w-full grid-cols-[1fr_auto_1fr] items-center gap-4">
            <Logo />
            <h1 className="max-w-[40vw] truncate text-center text-base font-semibold" title={regionName}>
                {regionName}
            </h1>
            <div className='flex justify-end gap-2'>
                <Export graphId={graphId} />
                <Button className={"h-10"} variant={"secondary"}>
                    <User />
                    Войти
                </Button>
            </div>
        </div>
    )
}
