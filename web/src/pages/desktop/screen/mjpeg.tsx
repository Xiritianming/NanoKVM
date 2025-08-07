import { useEffect } from 'react';
import { Image } from 'antd';
import clsx from 'clsx';
import { useAtomValue } from 'jotai';

import MonitorXIcon from '@/assets/images/monitor-x.svg';
import { stopFrameDetect } from '@/api/stream.ts';
import { getFrameDetect } from '@/lib/localstorage.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { mouseStyleAtom } from '@/jotai/mouse.ts';
import { resolutionAtom } from '@/jotai/screen.ts';

export const Mjpeg = () => {
  const resolution = useAtomValue(resolutionAtom);
  const mouseStyle = useAtomValue(mouseStyleAtom);

  useEffect(() => {
    // stop frame detect for a while
    const enabled = getFrameDetect();
    if (enabled) {
      stopFrameDetect(10);
    }
  }, [resolution]);

  return (
    <div className="flex h-screen w-screen items-start justify-center xl:items-center">
      <Image
        id="screen"
        className={clsx('block min-h-[480px] min-w-[640px] select-none', mouseStyle)}
        style={
          resolution?.width
            ? { width: resolution.width, height: resolution.height, objectFit: 'cover' }
            : { 
                maxWidth: '100vw', 
                maxHeight: '100vh', 
                objectFit: 'contain',
                aspectRatio: 'auto'
              }
        }
        src={`${getBaseUrl('http')}/api/stream/mjpeg`}
        fallback={MonitorXIcon}
        preview={false}
      />
    </div>
  );
};
