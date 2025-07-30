import { SpaceNetBrowser } from '@/components/SpaceNetBrowser';

export default function Home() {
  return (
    <SpaceNetBrowser
      serverAddr="::1"
      httpPort={8080}
      udpPort={1337}
      playerName="Anonymous"
    />
  );
}
