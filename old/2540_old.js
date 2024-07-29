  searchValue: string,
  options: Array<any>
) => {
  return filter(options, (option) => {
    const { label, detail, value } = option;
    const regex = new RegExp(escapeRegExp(searchValue), 'i');
    return regex.test(label) || regex.test(detail) || regex.test(value);
  });
};

export default class ItemsDropdown extends Component<ItemDropdownProps> {
  static defaultProps = {
    optionRenderer: (optionProps: ItemDropdown) => (
      <ItemDropdownOption {...optionProps} />
    ),
    selectionRenderer: (optionProps: ItemDropdown) => (
      <ItemDropdownOption selected {...optionProps} />
    ),
    onSearch: onSearchItemsDropdown,
    skin: SelectSkin,
  };
  render() {
    return <Select {...this.props} optionHeight={50} />;
  }
}
