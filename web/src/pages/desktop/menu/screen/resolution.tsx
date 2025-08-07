import { Popover, Tooltip } from 'antd';
import { useAtom } from 'jotai';
import { CheckIcon, CircleHelpIcon, RatioIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { updateScreen, getScreen } from '@/api/vm';
import { Resolution as TypeResolution } from '@/types';
import { setResolution as setCookie } from '@/lib/localstorage';
import { resolutionAtom } from '@/jotai/screen.ts';

const resolutions: TypeResolution[] = [
  { width: 0, height: 0 },
  { width: 1920, height: 1080 }, // 16:9
  { width: 1600, height: 900 },  // 16:9
  { width: 1280, height: 720 },  // 16:9
  { width: 1280, height: 960 },  // 4:3
  { width: 1024, height: 768 },  // 4:3
  { width: 800, height: 600 },   // 4:3
  { width: 640, height: 480 }    // 4:3
];

export const Resolution = () => {
  const { t } = useTranslation();
  const [resolution, setResolution] = useAtom(resolutionAtom);

  async function update(item: TypeResolution) {
    const rsp = await updateScreen('resolution', item.height);
    if (rsp.code !== 0) {
      return;
    }

    setResolution(item);
    setCookie(item);
  }

  const content = (
    <>
      {resolutions.map((res) => (
        <div
          key={res.height}
          className="flex cursor-pointer select-none items-center rounded py-1.5 pl-1 pr-5 hover:bg-neutral-700/70"
          onClick={() => update(res)}
        >
          <div className="flex h-[14px] w-[20px] items-end text-blue-500">
            {res.height === resolution?.height && <CheckIcon size={15} />}
          </div>

          {res.height === 0 ? (
            <div className="flex items-center justify-between space-x-2">
              <span>{t('screen.auto')}</span>
              <Tooltip
                title={t('screen.autoTips')}
                placement="right"
                overlayInnerStyle={{ width: '300px' }}
              >
                <CircleHelpIcon size={14} />
              </Tooltip>
            </div>
          ) : (
            <>
              <span>{res.width}</span>
              <span className="px-1">x</span>
              <span>{res.height}</span>
            </>
          )}
        </div>
      ))}
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <RatioIcon size={18} />
        <span className="select-none text-sm">{t('screen.resolution')}</span>
      </div>
    </Popover>
  );
};
