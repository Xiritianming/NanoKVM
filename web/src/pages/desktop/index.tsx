import { useEffect } from 'react';
import { useAtom, useAtomValue } from 'jotai';
import { useTranslation } from 'react-i18next';
import { useMediaQuery } from 'react-responsive';

import * as storage from '@/lib/localstorage.ts';
import { client } from '@/lib/websocket.ts';
import { isKeyboardEnableAtom } from '@/jotai/keyboard.ts';
import { resolutionAtom, videoModeAtom } from '@/jotai/screen.ts';
import { Head } from '@/components/head.tsx';

import { Keyboard } from './keyboard';
import { VirtualKeyboard } from './keyboard/virtual-keyboard';
import { Menu } from './menu';
import { Mouse } from './mouse';
import { Notification } from './notification.tsx';
import { Screen } from './screen';

export const Desktop = () => {
  const { t } = useTranslation();
  const isBigScreen = useMediaQuery({ minWidth: 850 });

  const [videoMode, setVideoMode] = useAtom(videoModeAtom);
  const [resolution, setResolution] = useAtom(resolutionAtom);
  const isKeyboardEnable = useAtomValue(isKeyboardEnableAtom);

  useEffect(() => {
    const mode = getVideoMode();
    setVideoMode(mode);

    const res = storage.getResolution() || { width: 0, height: 0 };
    setResolution(res);

    // Register for resolution change notifications
    client.register('resolution_change', (message) => {
      try {
        const data = JSON.parse(message.data as string);
        if (data.type === 'resolution_change') {
          // Only update if current resolution is auto mode (0x0)
          const currentResolution = storage.getResolution() || { width: 0, height: 0 };
          if (currentResolution.width === 0 && currentResolution.height === 0) {
            const newResolution = { width: data.width, height: data.height };
            setResolution(newResolution);
            console.log('Auto-updated resolution to:', newResolution);
          }
        }
      } catch (err) {
        console.error('Failed to parse resolution change message:', err);
      }
    });

    const timer = setInterval(() => {
      client.send([0]);
    }, 60 * 1000);

    return () => {
      clearInterval(timer);
      client.unregister('resolution_change');
      client.unregister('stream');
      client.close();
    };
  }, []);

  function getVideoMode() {
    const defaultVideoMode = window.RTCPeerConnection ? 'h264' : 'mjpeg';

    const cookieVideoMode = storage.getVideoMode();
    if (cookieVideoMode) {
      if (cookieVideoMode === 'direct' && !window.VideoDecoder) {
        return defaultVideoMode;
      }
      return cookieVideoMode;
    }

    return defaultVideoMode;
  }

  return (
    <>
      <Head title={t('head.desktop')} />

      {isBigScreen && <Notification />}

      {videoMode && resolution && (
        <>
          <Menu />
          <Screen />
          <Mouse />
          {isKeyboardEnable && <Keyboard />}
        </>
      )}

      <VirtualKeyboard />
    </>
  );
};
