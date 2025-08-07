import { useEffect, useState } from 'react';
import { Spin } from 'antd';
import clsx from 'clsx';
import { useAtomValue } from 'jotai';
import { w3cwebsocket as W3cWebSocket } from 'websocket';

import { getBaseUrl } from '@/lib/service.ts';
import { mouseStyleAtom } from '@/jotai/mouse.ts';
import { resolutionAtom } from '@/jotai/screen.ts';

export const H264Webrtc = () => {
  const resolution = useAtomValue(resolutionAtom);
  const mouseStyle = useAtomValue(mouseStyleAtom);

  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const videoElement = document.getElementById('screen') as HTMLVideoElement;

    const url = `${getBaseUrl('ws')}/api/stream/h264`;
    const ws = new W3cWebSocket(url);

    const pc = new RTCPeerConnection({
      iceServers: [
        {
          urls: ['stun:stun.l.google.com:19302']
        }
      ]
    });

    pc.ontrack = function (event) {
      if (event.track.kind !== 'video') {
        console.log('unhandled track kind: ', event.track.kind);
        return;
      }
      videoElement.srcObject = event.streams[0];
    };

    ws.onopen = () => {
      pc.onicecandidate = (event) => {
        if (event.candidate) {
          ws.send(JSON.stringify({ event: 'candidate', data: JSON.stringify(event.candidate) }));
        }
      };

      pc.addTransceiver('video', { direction: 'recvonly' });

      pc.createOffer({ offerToReceiveVideo: true })
        .then((offer) => {
          pc.setLocalDescription(offer).catch(console.log);
          ws.send(JSON.stringify({ event: 'offer', data: JSON.stringify(offer) }));
        })
        .catch(console.log);
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data as string);
        if (!msg?.data) return;

        const data = JSON.parse(msg.data);
        if (!data) return;

        switch (msg.event) {
          case 'answer':
            pc.setRemoteDescription(data).catch(console.log);
            break;

          case 'candidate':
            pc.addIceCandidate(data).catch(console.log);
            break;

          default:
            console.log('unhandled event: ', msg.event);
        }
      } catch (err) {
        console.log(err);
      }
    };

    const heartbeatTimer = setInterval(() => {
      try {
        ws.send(JSON.stringify({ event: 'heartbeat', data: '' }));
      } catch (err) {
        console.log(err);
      }
    }, 60 * 1000);

    setTimeout(() => {
      setIsLoading(false);
    }, 5 * 1000);

    return () => {
      ws.close();
      pc.close();
      if (heartbeatTimer) {
        clearInterval(heartbeatTimer);
      }
    };
  }, []);

  return (
    <Spin size="large" tip="Loading" spinning={isLoading}>
      <div className="flex h-screen w-screen items-start justify-center xl:items-center">
        <video
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
          muted
          autoPlay
          playsInline
          controls={false}
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
          }}
          onPlaying={() => {
            setIsLoading(false);
          }}
        />
      </div>
    </Spin>
  );
};
