import { world, BlockLocation, GameMode } from "mojang-minecraft";
import * as GameTest from "mojang-gametest";

const dim = world.getDimension("overworld");

function sendMessage(message) {
  dim.runCommand(`tell @a "${message}"`);
}

world.events.dataDrivenEntityTriggerEvent.subscribe((eventData) => {
  var event_name = eventData.id;
  // Entity object, see: https://docs.microsoft.com/en-us/minecraft/creator/scriptapi/mojang-minecraft/entity
  var entity = eventData.entity;

  // Ignore other events
  // We can also filter out things when calling .subscribe(), but for now do it this way
  if (event_name != "special_death_event" && event_name != "special_death_event_diamond") {
    return;
  }
  
  // Get JSON event from tag
  var eventb64 = entity.getTags()[0];
  var prefix = "[SIEMCRAFT]";
  if (!eventb64.startsWith(prefix)) {
    return;
  }
  eventb64 = eventb64.slice(prefix.length);

  // Send message as a JSON string
  var weapon = "other_weapon";
  if (event_name == "special_death_event_diamond") {
    weapon = "diamond_sword";
  }
  var message = prefix + `{"eventb64": "${eventb64}", "item": "${weapon}", "projectile": "" }`;
  sendMessage(message);
});
