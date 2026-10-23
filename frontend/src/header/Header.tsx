import { Export } from '@/components/export/Export'
import { Logo } from './Logo'
import { Button } from '@/components/ui/button'
import { User } from 'lucide-react'

type HeaderProps = {
    graphId?: string;
};

export function Header({ graphId }: HeaderProps) {
    return (
        <div className="flex w-full items-center justify-between">
            <Logo />
            <div className='flex gap-2'>
                <Export graphId={graphId} />
                <Button className={"h-10"} variant={"secondary"}>
                    <User />
                    Войти
                </Button>
            </div>
        </div>
    )
}
