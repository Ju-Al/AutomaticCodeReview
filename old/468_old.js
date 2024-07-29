    if (cell.cellType === 'markdown') {
      store.dispatch(actions.evaluateCell(cell.id))
    }
  })
  window.setTimeout(
    () => {
      cells.forEach((cell) => {
        if (cell.cellType !== 'markdown') {
          store.dispatch(actions.evaluateCell(cell.id))
        }
      })
    },
    42, // wait a few milliseconds to let React DOM updates flush
  )
}
