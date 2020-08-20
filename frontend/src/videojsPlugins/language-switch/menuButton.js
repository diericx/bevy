import videojs from 'video.js';
import LanguageMenuItem from './menuItem.js';

const MenuButton = videojs.getComponent('MenuButton');
let availableLanguages;
let defaultLanguage;

class LanguageMenuButton extends MenuButton {
  constructor(player, options) {

    availableLanguages = options.languages;
    defaultLanguage = options.defaultLanguage;

    super(player, options);
    this.addClass('vjs-language-switch');
    this.controlText(player.localize('Switch language'));

    this.options = options;
    this.addCustomIconClass(options);
  }

  addCustomIconClass(options) {
    const iconPlaceholder = this.el_.childNodes[0].children[0];
    const iconClass = options.buttonClass || 'icon-globe';

    iconPlaceholder.className += ' ' + iconClass;
  }

  createItems() {
    let menuItems = [];
    const player = this.player_;

    menuItems = availableLanguages.map(language => {
      return new LanguageMenuItem(player, {
        sources: language.sources,
        label: language.name,
        defaultSelection: defaultLanguage === language.name
      });
    });

    return menuItems;
  }
}

export default LanguageMenuButton;
