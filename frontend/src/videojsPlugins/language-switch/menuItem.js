import videojs from 'video.js';
const MenuItem = videojs.getComponent('MenuItem');

class LanguageMenuItem extends MenuItem {
  constructor(player, options) {
    super(player, options);

    this.addClass('vjs-language-switch__item');
    this.selectable = true;
    this.options = options;
    this.selected(options.defaultSelection);
  }

  handleClick(event) {
    this.player_.trigger('changedlanguage', this.options);
    this.selected(this.player_.activeLanguage === this.options.label);
  }
}

MenuItem.registerComponent('LanguageMenuItem', LanguageMenuItem);

export default LanguageMenuItem;
