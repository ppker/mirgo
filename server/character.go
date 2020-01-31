package main

import (
	"sync"
	"time"

	"github.com/yenkeia/mirgo/common"
	"github.com/yenkeia/mirgo/setting"
)

type Character struct {
	Player             *Player
	HP                 uint16
	MP                 uint16
	Level              uint16
	Experience         int64
	Gold               uint64
	GuildName          string
	GuildRankName      string
	Class              common.MirClass
	Gender             common.MirGender
	Hair               uint8
	Light              uint8
	Inventory          []common.UserItem // 46
	Equipment          []common.UserItem // 14
	QuestInventory     []common.UserItem // 40
	Trade              []common.UserItem // 10
	Refine             []common.UserItem // 16
	LooksArmour        int
	LooksWings         int
	LooksWeapon        int
	LooksWeaponEffect  int
	SendItemInfo       []common.ItemInfo
	CurrentBagWeight   int
	MaxHP              uint16
	MaxMP              uint16
	MinAC              uint16 // 物理防御力
	MaxAC              uint16
	MinMAC             uint16 // 魔法防御力
	MaxMAC             uint16
	MinDC              uint16 // 攻击力
	MaxDC              uint16
	MinMC              uint16 // 魔法力
	MaxMC              uint16
	MinSC              uint16 // 道术力
	MaxSC              uint16
	MaxExperience      int64
	Accuracy           uint8
	Agility            uint8
	CriticalRate       uint8
	CriticalDamage     uint8
	MaxBagWeight       uint16 //Other Stats;
	MaxWearWeight      uint16
	MaxHandWeight      uint16
	ASpeed             int8
	Luck               int8
	LifeOnHit          uint8
	HpDrainRate        uint8
	Reflect            uint8
	MagicResist        uint8
	PoisonResist       uint8
	HealthRecovery     uint8
	SpellRecovery      uint8
	PoisonRecovery     uint8
	Holy               uint8
	Freezing           uint8
	PoisonAttack       uint8
	ExpRateOffset      float32
	ItemDropRateOffset float32
	MineRate           uint8
	GemRate            uint8
	FishRate           uint8
	CraftRate          uint8
	GoldDropRateOffset float32
	AttackBonus        uint8
	Magics             []common.UserMagic
	ActionList         *sync.Map // map[uint32]DelayedAction
}

func NewCharacter(g *Game, p *Player, c *common.Character) Character {
	userItemIDIndexMap := make(map[int]int)
	cui := make([]common.CharacterUserItem, 0, 100)
	g.DB.Table("character_user_item").Where("character_id = ?", c.ID).Find(&cui)
	is := make([]int, 0, 46)
	es := make([]int, 0, 14)
	qs := make([]int, 0, 40)
	for _, i := range cui {
		switch common.UserItemType(i.Type) {
		case common.UserItemTypeInventory:
			is = append(is, i.UserItemID)
		case common.UserItemTypeEquipment:
			es = append(es, i.UserItemID)
		case common.UserItemTypeQuestInventory:
			qs = append(qs, i.UserItemID)
		}
		userItemIDIndexMap[i.UserItemID] = i.Index
	}
	inventory := make([]common.UserItem, 46)
	equipment := make([]common.UserItem, 14)
	questInventory := make([]common.UserItem, 40)
	trade := make([]common.UserItem, 0)
	refine := make([]common.UserItem, 0)
	uii := make([]common.UserItem, 0, 46)
	uie := make([]common.UserItem, 0, 14)
	uiq := make([]common.UserItem, 0, 40)
	g.DB.Table("user_item").Where("id in (?)", is).Find(&uii)
	g.DB.Table("user_item").Where("id in (?)", es).Find(&uie)
	g.DB.Table("user_item").Where("id in (?)", qs).Find(&uiq)
	for _, v := range uii {
		inventory[userItemIDIndexMap[int(v.ID)]] = v
	}
	for _, v := range uie {
		equipment[userItemIDIndexMap[int(v.ID)]] = v
	}
	for _, v := range uiq {
		questInventory[userItemIDIndexMap[int(v.ID)]] = v
	}
	magics := make([]common.UserMagic, 0)
	g.DB.Table("user_magic").Where("character_id = ?", c.ID).Find(&magics)
	return Character{
		Player:         p,
		HP:             c.HP,
		MP:             c.MP,
		Level:          c.Level,
		Experience:     c.Experience,
		Gold:           c.Gold,
		GuildName:      "", // TODO
		GuildRankName:  "", // TODO
		Class:          c.Class,
		Gender:         c.Gender,
		Hair:           c.Hair,
		Inventory:      inventory,
		Equipment:      equipment,
		QuestInventory: questInventory,
		Trade:          trade,
		Refine:         refine,
		SendItemInfo:   make([]common.ItemInfo, 0),
		MaxExperience:  100,
		Magics:         magics,
		ActionList:     new(sync.Map),
	}
}

func (c *Character) NewObjectID() uint32 {
	return c.Player.Map.Env.NewObjectID()
}

func (c *Character) IsDead() bool {
	return false
}

func (c *Character) IsHidden() bool {
	return false
}

func (c *Character) CanMove() bool {
	return true
}

func (c *Character) CanWalk() bool {
	return true
}

func (c *Character) CanRun() bool {
	return true
}

func (c *Character) CanAttack() bool {
	return true
}

func (c *Character) CanRegen() bool {
	return true
}

func (c *Character) CanCast() bool {
	return true
}

func (c *Character) CanUseItem(item *common.UserItem) bool {
	return true
}

func (c *Character) EnqueueItemInfos() {
	gdb := c.Player.Map.Env.GameDB
	itemInfos := make([]*common.ItemInfo, 0)
	for i := range c.Inventory {
		itemID := int(c.Inventory[i].ItemID)
		if itemID == 0 {
			continue
		}
		itemInfos = append(itemInfos, gdb.GetItemInfoByID(itemID))
	}
	for i := range c.Equipment {
		itemID := int(c.Equipment[i].ItemID)
		if itemID == 0 {
			continue
		}
		itemInfos = append(itemInfos, gdb.GetItemInfoByID(itemID))
	}
	for i := range c.QuestInventory {
		itemID := int(c.QuestInventory[i].ItemID)
		if itemID == 0 {
			continue
		}
		itemInfos = append(itemInfos, gdb.GetItemInfoByID(itemID))
	}
	for i := range itemInfos {
		c.EnqueueItemInfo(itemInfos[i].ID)
	}
}

func (c *Character) EnqueueItemInfo(itemID int32) {
	for m := range c.SendItemInfo {
		s := c.SendItemInfo[m]
		if s.ID == itemID {
			return
		}
	}
	item := c.Player.Map.Env.GameDB.GetItemInfoByID(int(itemID))
	if item == nil {
		return
	}
	c.Player.Enqueue(ServerMessage{}.NewItemInfo(item))
	c.SendItemInfo = append(c.SendItemInfo, *item)
}

func (c *Character) EnqueueQuestInfo() {

}

func (c *Character) RefreshStats() {
	c.RefreshLevelStats()
	c.RefreshBagWeight()
	c.RefreshEquipmentStats()
	c.RefreshItemSetStats()
	c.RefreshMirSetStats()
	c.RefreshSkills()
	c.RefreshBuffs()
	c.RefreshStatCaps()
	c.RefreshMountStats()
	c.RefreshGuildBuffs()
}

func (c *Character) RefreshLevelStats() {
	baseStats := setting.BaseStats[c.Class]
	c.Accuracy = uint8(baseStats.StartAccuracy)
	c.Agility = uint8(baseStats.StartAgility)
	c.CriticalRate = uint8(baseStats.StartCriticalRate)
	c.CriticalDamage = uint8(baseStats.StartCriticalDamage)
	c.MaxExperience = 100
	c.MaxHP = uint16(14 + (float32(c.Level)/baseStats.HpGain+baseStats.HpGainRate)*float32(c.Level))
	c.MinAC = 0
	if baseStats.MinAc > 0 {
		c.MinAC = uint16(int(c.Level) / baseStats.MinAc)
	}
	c.MaxAC = 0
	if baseStats.MaxAc > 0 {
		c.MaxAC = uint16(int(c.Level) / baseStats.MaxAc)
	}
	c.MinMAC = 0
	if baseStats.MinMac > 0 {
		c.MinMAC = uint16(int(c.Level) / baseStats.MinMac)
	}
	c.MaxMAC = 0
	if baseStats.MaxMac > 0 {
		c.MaxMAC = uint16(int(c.Level) / baseStats.MaxMac)
	}
	c.MinDC = 0
	if baseStats.MinDc > 0 {

		c.MinDC = uint16(int(c.Level) / baseStats.MinDc)
	}
	c.MaxDC = 0
	if baseStats.MaxDc > 0 {
		c.MaxDC = uint16(int(c.Level) / baseStats.MaxDc)
	}
	c.MinMC = 0
	if baseStats.MinMc > 0 {
		c.MinMC = uint16(int(c.Level) / baseStats.MinMc)
	}
	c.MaxMC = 0
	if baseStats.MaxMc > 0 {
		c.MaxMC = uint16(int(c.Level) / baseStats.MaxMc)
	}
	c.MinSC = 0
	if baseStats.MinSc > 0 {
		c.MinSC = uint16(int(c.Level) / baseStats.MinSc)
	}
	c.MaxSC = 0
	if baseStats.MaxSc > 0 {
		c.MaxSC = uint16(int(c.Level) / baseStats.MaxSc)
	}
	c.CriticalRate = 0
	if baseStats.CritialRateGain > 0 {
		c.CriticalRate = uint8(float32(c.CriticalRate) + (float32(c.Level) / baseStats.CritialRateGain))
	}
	c.CriticalDamage = 0
	if baseStats.CriticalDamageGain > 0 {
		c.CriticalDamage = uint8(float32(c.CriticalDamage) + (float32(c.Level) / baseStats.CriticalDamageGain))
	}
	c.MaxBagWeight = uint16(50.0 + float32(c.Level)/baseStats.BagWeightGain*float32(c.Level))
	c.MaxWearWeight = uint16(15.0 + float32(c.Level)/baseStats.WearWeightGain*float32(c.Level))
	c.MaxHandWeight = uint16(12.0 + float32(c.Level)/baseStats.HandWeightGain*float32(c.Level))
	switch c.Class {
	case common.MirClassWarrior:
		c.MaxHP = uint16(14.0 + (float32(c.Level)/baseStats.HpGain+baseStats.HpGainRate+float32(c.Level)/20.0)*float32(c.Level))
		c.MaxMP = uint16(11.0 + (float32(c.Level) * 3.5) + (float32(c.Level) * baseStats.MpGainRate))
	case common.MirClassWizard:
		c.MaxMP = uint16(13.0 + (float32(c.Level/5.0+2.0) * 2.2 * float32(c.Level)) + (float32(c.Level) * baseStats.MpGainRate))
	case common.MirClassTaoist:
		c.MaxMP = uint16((13 + float32(c.Level)/8.0*2.2*float32(c.Level)) + (float32(c.Level) * baseStats.MpGainRate))
	}
}

func (c *Character) RefreshBagWeight() {
	c.CurrentBagWeight = 0
	for i := range c.Inventory {
		ui := c.Inventory[i]
		if ui.ID != 0 {
			it := c.Player.Map.Env.GameDB.GetItemInfoByID(int(ui.ItemID))
			c.CurrentBagWeight += int(it.Weight)
		}
	}
}

func (c *Character) RefreshEquipmentStats() {
	gdb := c.Player.Map.Env.GameDB
	for i := range c.Equipment {
		e := gdb.GetItemInfoByID(int(c.Equipment[i].ItemID))
		if e == nil {
			continue
		}
		switch e.Type {
		case common.ItemTypeArmour:
			c.LooksArmour = int(e.Shape)
			c.LooksWings = int(e.Effect)
		case common.ItemTypeWeapon:
			c.LooksWeapon = int(e.Shape)
			c.LooksWeaponEffect = int(e.Effect)
		}
	}
}

func (c *Character) RefreshItemSetStats() {

}

func (c *Character) RefreshMirSetStats() {

}

func (c *Character) RefreshSkills() {

}

func (c *Character) RefreshBuffs() {

}

func (c *Character) RefreshStatCaps() {

}

func (c *Character) RefreshMountStats() {

}

func (c *Character) RefreshGuildBuffs() {

}

// GetUserItemByID 获取物品，返回该物品在容器的索引和是否成功
func (c *Character) GetUserItemByID(mirGridType common.MirGridType, id uint64) (index int, item *common.UserItem) {
	var arr []common.UserItem
	switch mirGridType {
	case common.MirGridTypeInventory:
		arr = c.Inventory
	case common.MirGridTypeEquipment:
		arr = c.Equipment
	default:
		panic("error mirGridType")
	}
	for i := range arr {
		item := arr[i]
		if item.ID == id {
			return i, &item
		}
	}
	return -1, nil
}

// GainItem 为玩家增加物品，增加成功返回 true
func (c *Character) GainItem(ui *common.UserItem) bool {
	item := c.Player.Map.Env.GameDB.GetItemInfoByID(int(ui.ItemID))
	if item == nil {
		return false
	}
	i, j := 6, 46
	if item.Type == common.ItemTypePotion ||
		item.Type == common.ItemTypeScroll ||
		item.Type == common.ItemTypeScript ||
		item.Type == common.ItemTypeAmulet {
		i = 0
		j = 4
	} else if item.Type == common.ItemTypeAmulet {
		i = 4
		j = 6
	} else {
		i = 6
	}
	for i < j {
		if c.Inventory[i].ID != 0 {
			i++
			continue
		}
		c.Inventory[i] = *ui
		break
	}
	c.EnqueueItemInfo(ui.ItemID)
	c.Player.Enqueue(ServerMessage{}.GainedItem(ui))
	c.RefreshBagWeight()
	return true
}

// GainGold 为玩家增加金币
func (c *Character) GainGold(gold uint64) {
	if gold <= 0 {
		return
	}
	c.Gold += gold
	c.Player.Enqueue(ServerMessage{}.GainedGold(gold))
}

func (c *Character) UpdateConcentration() {
	c.Player.Enqueue(ServerMessage{}.SetConcentration(c.Player))
	c.Player.Broadcast(ServerMessage{}.SetObjectConcentration(c.Player))
}

func (c *Character) GetAttackPower(min, max int) int {
	if min < 0 {
		min = 0
	}
	if max < min {
		max = min
	}
	// TODO luck
	return G_Rand.RandInt(min, max+1)
}

// TODO
func (c *Character) Attacked(attacker IMapObject, damageFinal int, defenceType common.DefenceType, damageWeapon bool) {

}

func (c *Character) GainExp(amount uint32) {
	c.Experience += int64(amount)
	c.Player.Enqueue(ServerMessage{}.GainExperience(amount))
	if c.Experience < c.MaxExperience {
		return
	}
	c.Experience -= c.MaxExperience
	c.Level++
	c.LevelUp()
}

func (c *Character) SetHP(amount uint32) {
	c.HP = uint16(amount)
	msg := ServerMessage{}.HealthChanged(c.HP, c.MP)
	c.Player.Enqueue(msg)
	c.Player.Broadcast(msg)
}

func (c *Character) SetMP(amount uint32) {
	c.MP = uint16(amount)
	msg := ServerMessage{}.HealthChanged(c.HP, c.MP)
	c.Player.Enqueue(msg)
	c.Player.Broadcast(msg)
}

func (c *Character) ChangeHP(amount int) {
	if amount == 0 || c.IsDead() {
		return
	}
	c.SetHP(uint32(int(c.HP) + amount))
}

func (c *Character) ChangeMP(amount int) {
	if amount == 0 || c.IsDead() {
		return
	}
	c.SetMP(uint32(int(c.MP) + amount))
}

func (c *Character) LevelUp() {
	c.RefreshStats()
	c.SetHP(uint32(c.MaxHP))
	c.SetMP(uint32(c.MaxMP))
	c.Player.Enqueue(ServerMessage{}.LevelChanged(c.Level, c.Experience, c.MaxExperience))
	c.Player.Broadcast(ServerMessage{}.ObjectLeveled(c.Player.GetID()))
}

func (c *Character) Process() {
	finishID := make([]uint32, 0)
	c.ActionList.Range(func(k, v interface{}) bool {
		action := v.(*DelayedAction)
		if action.Finish || time.Now().Before(action.ActionTime) {
			return true
		}
		action.Task.Execute()
		action.Finish = true
		if action.Finish {
			finishID = append(finishID, action.ID)
		}
		return true
	})
	for i := range finishID {
		c.ActionList.Delete(finishID[i])
	}
}

func (c *Character) GetMagic(spell common.Spell) *common.UserMagic {
	for i := range c.Magics {
		userMagic := c.Magics[i]
		if userMagic.Spell == spell {
			return &userMagic
		}
	}
	return nil
}

func (c *Character) GetClientMagics() []common.ClientMagic {
	gdb := c.Player.Map.Env.GameDB
	res := make([]common.ClientMagic, 0)
	for i := range c.Magics {
		userMagic := c.Magics[i]
		info := gdb.GetMagicInfoByID(userMagic.MagicID)
		res = append(res, userMagic.GetClientMagic(info))
	}
	return res
}

func (c *Character) UseMagic(spell common.Spell, magic *common.UserMagic, target IMapObject) (cast bool, targetID uint32) {
	cast = true
	switch spell {
	case common.SpellFireBall, common.SpellGreatFireBall, common.SpellFrostCrunch:
		if ok := c.Fireball(target, magic); !ok {
			targetID = 0
		}
	case common.SpellHealing:
		if target == nil {
			target = c.Player
			targetID = c.Player.GetID()
		}
		c.Healing(target, magic)
	case common.SpellRepulsion, common.SpellEnergyRepulsor, common.SpellFireBurst:
		c.Repulsion(magic)
	case common.SpellElectricShock:
		// ActionList.Add(new DelayedAction(DelayedType.Magic, Envir.Time + 500, magic, target as MonsterObject));
		action := NewDelayedAction(c.NewObjectID(), DelayedTypeMagic, NewTask(c.CompleteMagic, magic, target))
		c.ActionList.Store(action.ID, action)
	case common.SpellPoisoning:
		if !c.Poisoning(target, magic) {
			cast = false
		}
	case common.SpellHellFire:
		c.HellFire(magic)
	case common.SpellThunderBolt:
		c.ThunderBolt(target, magic)
	case common.SpellSoulFireBall:
		// if (!SoulFireball(target, magic, out cast)) targetID = 0;
		if !c.SoulFireball(target, magic) {
			targetID = 0
			cast = false
		}
	case common.SpellSummonSkeleton:
		c.SummonSkeleton(magic)
	case common.SpellTeleport, common.SpellBlink:
		// ActionList.Add(new DelayedAction(DelayedType.Magic, Envir.Time + 200, magic, location));
		action := NewDelayedAction(c.NewObjectID(), DelayedTypeMagic, NewTask(c.CompleteMagic, magic, c.Player.GetPoint()))
		c.ActionList.Store(action.ID, action)
	case common.SpellHiding:
		c.Hiding(magic)
	case common.SpellHaste, common.SpellLightBody:
		// ActionList.Add(new DelayedAction(DelayedType.Magic, Envir.Time + 500, magic));
		action := NewDelayedAction(c.NewObjectID(), DelayedTypeMagic, NewTask(c.CompleteMagic, magic))
		c.ActionList.Store(action.ID, action)
	case common.SpellFury:
		cast = c.FurySpell(magic)
	case common.SpellImmortalSkin:
		cast = c.ImmortalSkin(magic)
	case common.SpellFireBang, common.SpellIceStorm:
		// FireBang(magic, target == null ? location : target.CurrentLocation);
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		c.FireBang(magic, location)
	case common.SpellMassHiding:
		// MassHiding(magic, target == null ? location : target.CurrentLocation, out cast);
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		cast = c.MassHiding(magic, location)
	case common.SpellSoulShield, common.SpellBlessedArmour:
		// SoulShield(magic, target == null ? location : target.CurrentLocation, out cast);
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		cast = c.SoulShield(magic, location)
	case common.SpellFireWall:
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		c.FireWall(magic, location)
	case common.SpellLightning:
		c.Lightning(magic)
	case common.SpellHeavenlySword:
		c.HeavenlySword(magic)
	case common.SpellMassHealing:
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		c.MassHealing(magic, location)
	case common.SpellShoulderDash:
		c.ShoulderDash(magic)
	case common.SpellThunderStorm, common.SpellFlameField, common.SpellStormEscape:
		/*
			ThunderStorm(magic);
			if (spell == Spell.FlameField)
				SpellTime = Envir.Time + 2500; //Spell Delay
			if (spell == Spell.StormEscape)
				//Start teleport.
				ActionList.Add(new DelayedAction(DelayedType.Magic, Envir.Time + 750, magic, location));
		*/
	case common.SpellMagicShield:
		// ActionList.Add(new DelayedAction(DelayedType.Magic, Envir.Time + 500, magic, magic.GetPower(GetAttackPower(MinMC, MaxMC) + 15)));
		action := NewDelayedAction(c.NewObjectID(), DelayedTypeMagic, NewTask(c.CompleteMagic, magic, magic.GetPower(c.GetAttackPower(int(c.MinMC), int(c.MaxMC))+15)))
		c.ActionList.Store(action.ID, action)
	case common.SpellFlameDisruptor:
		c.FlameDisruptor(target, magic)
	case common.SpellTurnUndead:
		c.TurnUndead(target, magic)
	case common.SpellMagicBooster:
		c.MagicBooster(magic)
	case common.SpellVampirism:
		c.Vampirism(target, magic)
	case common.SpellSummonShinsu:
		c.SummonShinsu(magic)
	case common.SpellPurification:
		/*
			if (target == null)
			{
				target = this;
				targetID = ObjectID;
			}
			Purification(target, magic);
		*/
	case common.SpellLionRoar, common.SpellBattleCry:
		// CurrentMap.ActionList.Add(new DelayedAction(DelayedType.Magic, Envir.Time + 500, this, magic, CurrentLocation));
	case common.SpellRevelation:
		c.Revelation(target, magic)
	case common.SpellPoisonCloud:
		cast = c.PoisonCloud(magic, c.Player.GetPoint())
	case common.SpellEntrapment:
		c.Entrapment(target, magic)
	case common.SpellBladeAvalanche:
		c.BladeAvalanche(magic)
	case common.SpellSlashingBurst:
		cast = c.SlashingBurst(magic)
	case common.SpellRage:
		c.Rage(magic)
	case common.SpellMirroring:
		c.Mirroring(magic)
	case common.SpellBlizzard:
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		cast = c.Blizzard(magic, location)
	case common.SpellMeteorStrike:
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		cast = c.MeteorStrike(magic, location)
	case common.SpellIceThrust:
		c.IceThrust(magic)
	case common.SpellProtectionField:
		c.ProtectionField(magic)
	case common.SpellPetEnhancer:
		cast = c.PetEnhancer(target, magic)
	case common.SpellTrapHexagon:
		cast = c.TrapHexagon(magic, target)
	case common.SpellReincarnation:
		// Reincarnation(magic, target == null ? null : target as PlayerObject, out cast);
		if target != nil {
			target = c.Player
		}
		cast = c.Reincarnation(magic, target)
	case common.SpellCurse:
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		cast = c.Curse(magic, location)
	case common.SpellSummonHolyDeva:
		c.SummonHolyDeva(magic)
	case common.SpellHallucination:
		c.Hallucination(target, magic)
	case common.SpellEnergyShield:
		cast = c.EnergyShield(target, magic)
	case common.SpellUltimateEnhancer:
		cast = c.UltimateEnhancer(target, magic)
	case common.SpellPlague:
		location := target.GetPoint()
		if target == nil {
			location = c.Player.GetPoint()
		}
		cast = c.Plague(magic, location)
	default:
		cast = false
	}
	return
}

func (c *Character) CompleteMagic(args ...interface{}) {
	userMagic := args[0].(*common.UserMagic)
	switch userMagic.Spell {
	// #region FireBall, GreatFireBall, ThunderBolt, SoulFireBall, FlameDisruptor
	case common.SpellFireBall, common.SpellGreatFireBall, common.SpellThunderBolt, common.SpellSoulFireBall, common.SpellFlameDisruptor, common.SpellStraightShot, common.SpellDoubleShot:
		value := args[1].(int)
		target := args[2].(IMapObject)
		if target == nil || !target.IsAttackTarget(c.Player) {
			return
		}
		if target.GetRace() == common.ObjectTypePlayer {
			target.(*Player).Attacked(c.Player, value, common.DefenceTypeMAC, false)
		} else if target.GetRace() == common.ObjectTypeMonster {
			target.(*Monster).Attacked(c.Player, value, common.DefenceTypeMAC, false)
		}
		return
	}
	// TODO #region FrostCrunch
	// TODO #region Vampirism
}

func (c *Character) CompleteAttack(args ...interface{})          {}
func (c *Character) CompleteMapMovement(args ...interface{})     {}
func (c *Character) CompleteMine(args ...interface{})            {}
func (c *Character) CompleteNPC(args ...interface{})             {}
func (c *Character) CompletePoison(args ...interface{})          {}
func (c *Character) CompleteDamageIndicator(args ...interface{}) {}

func (c *Character) Fireball(target IMapObject, magic *common.UserMagic) bool {
	if target == nil || !target.IsAttackTarget(c.Player) {
		return false
	}
	damage := magic.GetDamage(c.GetAttackPower(int(c.MinMC), int(c.MaxMC)))
	action := NewDelayedAction(c.NewObjectID(), DelayedTypeMagic, NewTask(c.CompleteMagic, magic, damage, target))
	c.ActionList.Store(action.ID, action)
	return true
}

func (c *Character) Healing(target IMapObject, magic *common.UserMagic) {
	if target == nil || !target.IsFriendlyTarget(c.Player) {
		return
	}
	// int health = magic.GetDamage(GetAttackPower(MinSC, MaxSC) * 2) + Level;
	health := magic.GetDamage(c.GetAttackPower(int(c.MinSC), int(c.MaxSC))*2) + int(c.Level)
	action := NewDelayedAction(c.NewObjectID(), DelayedTypeMagic, NewTask(c.CompleteMagic, magic, health, target))
	c.ActionList.Store(action.ID, action)
}

func (c *Character) Repulsion(magic *common.UserMagic) {

}

func (c *Character) Poisoning(target IMapObject, magic *common.UserMagic) bool {
	return true
}

func (c *Character) HellFire(magic *common.UserMagic) {

}

func (c *Character) ThunderBolt(target IMapObject, magic *common.UserMagic) {

}

func (c *Character) SoulFireball(target IMapObject, magic *common.UserMagic) bool {
	return true
}

func (c *Character) SummonSkeleton(magic *common.UserMagic)                           {}
func (c *Character) Hiding(magic *common.UserMagic)                                   {}
func (c *Character) FurySpell(magic *common.UserMagic) bool                           { return true }
func (c *Character) ImmortalSkin(magic *common.UserMagic) bool                        { return true }
func (c *Character) FireBang(magic *common.UserMagic, location common.Point)          {}
func (c *Character) MassHiding(magic *common.UserMagic, location common.Point) bool   { return true }
func (c *Character) SoulShield(magic *common.UserMagic, location common.Point) bool   { return true }
func (c *Character) FireWall(magic *common.UserMagic, location common.Point)          {}
func (c *Character) Lightning(magic *common.UserMagic)                                {}
func (c *Character) HeavenlySword(magic *common.UserMagic)                            {}
func (c *Character) MassHealing(magic *common.UserMagic, location common.Point)       {}
func (c *Character) ShoulderDash(magic *common.UserMagic)                             {}
func (c *Character) FlameDisruptor(target IMapObject, magic *common.UserMagic)        {}
func (c *Character) TurnUndead(target IMapObject, magic *common.UserMagic)            {}
func (c *Character) MagicBooster(magic *common.UserMagic)                             {}
func (c *Character) Vampirism(target IMapObject, magic *common.UserMagic)             {}
func (c *Character) SummonShinsu(magic *common.UserMagic)                             {}
func (c *Character) Revelation(target IMapObject, magic *common.UserMagic)            {}
func (c *Character) PoisonCloud(magic *common.UserMagic, location common.Point) bool  { return true }
func (c *Character) Entrapment(target IMapObject, magic *common.UserMagic)            {}
func (c *Character) BladeAvalanche(magic *common.UserMagic)                           {}
func (c *Character) SlashingBurst(magic *common.UserMagic) bool                       { return true }
func (c *Character) Rage(magic *common.UserMagic)                                     {}
func (c *Character) Mirroring(magic *common.UserMagic)                                {}
func (c *Character) Blizzard(magic *common.UserMagic, location common.Point) bool     { return true }
func (c *Character) MeteorStrike(magic *common.UserMagic, location common.Point) bool { return true }
func (c *Character) IceThrust(magic *common.UserMagic)                                {}
func (c *Character) ProtectionField(magic *common.UserMagic)                          {}
func (c *Character) PetEnhancer(target IMapObject, magic *common.UserMagic) bool      { return true }
func (c *Character) TrapHexagon(magic *common.UserMagic, target IMapObject) bool      { return true }
func (c *Character) Reincarnation(magic *common.UserMagic, target IMapObject) bool    { return true }
func (c *Character) Curse(magic *common.UserMagic, location common.Point) bool        { return true }
func (c *Character) SummonHolyDeva(magic *common.UserMagic)                           {}
func (c *Character) Hallucination(target IMapObject, magic *common.UserMagic) bool    { return true }
func (c *Character) EnergyShield(target IMapObject, magic *common.UserMagic) bool     { return true }
func (c *Character) UltimateEnhancer(target IMapObject, magic *common.UserMagic) bool { return true }
func (c *Character) Plague(magic *common.UserMagic, location common.Point) bool       { return true }
