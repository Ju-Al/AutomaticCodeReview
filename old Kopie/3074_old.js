import { Node } from 'tiptap';
          default: 'p'
import {
  setBlockType
} from 'tiptap-commands';

export default class Styles extends Node {

  get name() {
    return 'styles';
  }

  get schema() {
    const defaultStyle = (this.options.styles && this.options.styles[0]) || {
      tag: 'p'
    };
    return {
      attrs: {
        tag: {
          default: defaultStyle.p
        },
        class: {
          default: defaultStyle.class
        }
      },
      content: 'inline*',
      group: 'block',
      defining: true,
      draggable: false,

      parseDOM: (this.options.styles || []).map(style => {
        return {
          tag: style.class ? `${style.tag}[class="${style.class}"]` : `${style.tag}:not([class])`,
          attrs: {
            tag: style.tag,
            class: style.class
          }
        };
      }),

      toDOM: node => {
        const attrs = {};
        if (node.attrs.class) {
          attrs.class = node.attrs.class;
        }
        return [ node.attrs.tag, attrs, 0 ];
      }

    };
  }

  commands({ type, schema }) {
    return attrs => {
      return setBlockType(type, {
        class: null,
        ...attrs
      });
    };
  }
  // TODO: Clean up.
  // keys({ type }) {
  //   return this.options.levels.reduce((items, level) => ({
  //     ...items,
  //     ...{
  //       [`Shift-Ctrl-${level}`]: setBlockType(type, { level }),
  //     },
  //   }), {})
  // }

  // inputRules({ type }) {
  //   return this.options.levels.map(level => textblockTypeInputRule(
  //     new RegExp(`^(#{1,${level}})\\s$`),
  //     type,
  //     () => ({ level }),
  //   ))
  // }

}