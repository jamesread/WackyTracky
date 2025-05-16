export class DatabaseRemoteServer {
  async getTasks (theListId) {
    let jsonTasks = await window.client.listTasks({
      "parentId": theListId,
      "parentType": "list", 
    })

    window.dbal.local.addTasksFromServer(jsonTasks)
  }

  async getSubtasks(theTaskId) {
    let jsonTasks = await window.client.listTasks({
      "parentId": theTaskId,
      "parentType": "task",
    })

    window.dbal.local.addTasksFromServer(jsonTasks)
  }

  async fetchTags () {
    window.sidepanel.clearTags() // FIXME

    let json = await window.client.getTags({})

    window.dbal.local.addTagsFromServer(json)
  }

  async fetchLists () {
    let res = await window.client.getLists({})

    for (let l of res.lists) {
      window.dbal.local.addListFromServer(l)
    }
  }
}
