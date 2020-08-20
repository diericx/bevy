import videojs from 'video.js';
import LanguageMenuButton from './menuButton.js';

/**
* remove selected class form the options
*/

const unselectItems = (player) => {
  const items = player.el().getElementsByClassName('vjs-language-switch__item');

  Array.from(items).forEach((item) => {
    item.classList.remove('vjs-selected');
  });
};

/**
* event on selected the language
*/

const onLanguageSelect = (player, language, selected) => {

  let currentTime = player.currentTime();

  player.activeLanguage = language.label;

  unselectItems(player);

  const selectedCurrentSource = player.currentSources().filter((src) => {
    if (src.selected) {
      return src.selected === true;
    }
    return;
  });

  if (selectedCurrentSource.length) {
    const currenSelectedType = selectedCurrentSource[0].type;
    const currenSelectedQuality = selectedCurrentSource[0].label;

    /* move selected src to the first place in the array */
    /* because that the one which will be automaticaly played by the player */
    const orderedSouces = language.sources;

    language.sources.map((src, index) => {
      if (src.label === currenSelectedQuality && src.type === currenSelectedType) {
        const selectedSrc = orderedSouces.splice(index, 1);

        orderedSouces.unshift(selectedSrc[0]);
      }
    });
  }

  player.src(language.sources.map(function(src, index) {
    const defaultSrcData =
    { src: src.src, type: src.type, res: src.res, label: src.label };

    if (!selectedCurrentSource.length && index === 0) {
      return Object.assign(defaultSrcData, { selected: true });
    }

    return defaultSrcData;
  }));

  player.on('loadedmetadata', function() {
    player.currentTime(currentTime);
    player.play();
  });
};

/**
 * Function to invoke when the player is ready.
 * @function onPlayerReady
 * @param    {Player} player
 * @param    {Object} [options={}]
 */
const onPlayerReady = (player, options) => {
  player.on('changedlanguage', function(event, newSource) {
    onLanguageSelect(player, newSource);
  });

  player.getChild('controlBar')
  .addChild('LanguageMenuButton', options, options.positionIndex);
};

/**
 * A video.js plugin.
 *
 * In the plugin function, the value of `this` is a video.js `Player`
 * instance. You cannot rely on the player being in a 'ready' state here,
 * depending on how the plugin is invoked. This may or may not be important
 * to you; if not, remove the wait for 'ready'!
 *
 * @function languageSwitch
 * @param    {Object} [options={}]
 *           An object of options left to the plugin author to define.
 */
const languageSwitch = function(options) {
  this.ready(() => {
    onPlayerReady(this, options);
  });
};

videojs.registerComponent('LanguageMenuButton', LanguageMenuButton);

// Register the plugin with video.js.
videojs.registerPlugin('languageSwitch', languageSwitch);

// Include the version number.
languageSwitch.VERSION = '__VERSION__';

export default languageSwitch;
