import { DownloadIcon } from "lucide-react";
import { Button } from "../ui/button";

export function Export() {
    return (
        <>
            <Button className={"h-10"}>
                <DownloadIcon />
                Экспорт
            </Button>
        </>
    )
}