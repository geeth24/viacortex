import { cn } from "@/lib/utils";

interface LogoProps {
  className?: string;
}

function Logo({ className }: LogoProps) {
  return (
    <svg 
      width="42" 
      height="42" 
      viewBox="0 0 42 42" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={cn("text-secondary-foreground", className)}
    >
      <path 
        fillRule="evenodd" 
        clipRule="evenodd" 
        d="M22.0997 17.8596L31.9992 7.96015L29.8779 5.83883L22.5391 13.1777V0H19.5391V13.1777L12.2002 5.83883L10.0789 7.96015L19.9784 17.8596L21.0391 18.9203L22.0997 17.8596ZM24.2204 19.981L34.1199 10.0815L36.2412 12.2028L28.9024 19.5416H41.998V22.5416H28.9024L36.2412 29.8805L34.1199 32.0018L24.2204 22.1023L23.1598 21.0416L24.2204 19.981ZM7.9582 32.0018L17.8577 22.1023L18.9184 21.0416L17.8577 19.981L7.9582 10.0815L5.83688 12.2028L13.1757 19.5416H0V22.5416H13.1757L5.83688 29.8805L7.9582 32.0018ZM19.9784 24.2236L10.0789 34.1231L12.2002 36.2444L19.5391 28.9056V42H22.5391V28.9056L29.8779 36.2444L31.9992 34.1231L22.0997 24.2236L21.0391 23.1629L19.9784 24.2236Z" 
        fill="currentColor"
      />
    </svg>
  );
}

export default Logo;