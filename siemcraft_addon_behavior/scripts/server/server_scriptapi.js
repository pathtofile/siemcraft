var system = server.registerSystem(0,0);

system.initialize = function() {
  this.listenForEvent("minecraft:entity_death", (eventData) => this.onEntityDeath(eventData));
};

// Helper function to tell as message to be picked up by
// the websocket in a 'PlayerMessage' event
system.sendMessage = function (message) {
  let ExecuteEventData = this.createEventData("minecraft:execute_command");
  ExecuteEventData.data.command = "/tell @a \"" + message + "\"";
  this.broadcastEvent("minecraft:execute_command", ExecuteEventData);
};

system.onEntityDeath = function (eventData) {
  // Only look for entities killed by the player
  if (eventData.data.killer.__identifier__ != "minecraft:player") {
    return;
  }
  // Only look for entities killed with a name
  let killed_name = this.getComponent(eventData.data.entity, "minecraft:nameable");
  if (killed_name == null) {
    return;
  }

  // Get weapon in main hand, assume it was what was used to kill it
  let handContainer = system.getComponent(eventData.data.killer, "minecraft:hand_container");
  let mainHandItem = handContainer.data[0];

  // If there was a projectile, get that projectile_type
  let projectile = ""
  if (eventData.data.projectile_type != null && eventData.data.projectile_type !== "undefined") {
    projectile = eventData.data.projectile_type
  }

  // Get the raw event JSON we saved as a tag
  let tags = system.getComponent(eventData.data.entity, "minecraft:tag");
  if (tags == null) {
    return;
  }

  let eventb64 = tags.data[0];
  var prefix = "[SIEMCRAFT]";
  if (!eventb64.startsWith(prefix)) {
    return;
  }
  eventb64 = eventb64.slice(prefix.length);

  // Send message as a JSON string
  let message = prefix + '{"eventb64": "' + eventb64 + '",\n"item":"' + mainHandItem.item + '",\n"projectile": "' + projectile + '"}'
  this.sendMessage(message);
};
