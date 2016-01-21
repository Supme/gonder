package panel
import (
"regexp"
"strings"
)

func getGroups() map[string]string {
	var key, name string
	groups := make(map[string]string)
	query, err := Db.Query("SELECT `id`, `name` FROM `group`")
	checkErr(err)
	defer query.Close()
	for query.Next() {
		err = query.Scan(&key, &name)
		groups[key] = name
	}
	return groups
}

func addGroup(name string) {
	_, err := Db.Query("INSERT INTO `group` (`name`) VALUES (?)", name)
	checkErr(err)
}

func getCampaigns(id string) map[string]campaign {
	var key, name, subject string
	campaigns := make(map[string]campaign)
	query, err := Db.Query("SELECT `id`, `name`, `subject` FROM `campaign` WHERE group_id=?", id)
	checkErr(err)
	defer query.Close()
	for query.Next() {
		err = query.Scan(&key, &name, &subject)
		checkErr(err)
		campaigns[key] = campaign{Name: name, Subject: subject}
	}
	return campaigns
}

func addCampaigns(id string, name string) {
	_, err := Db.Query("INSERT INTO `campaign`(`group_id`, `interface_id`, `from`, `from_name`, `name`, `subject`, `body`, `start_time`, `end_time`) VALUES (?,0,'','',?,'New clear campaign','',NOW(),NOW())", id, name)
	checkErr(err)
}

func getCampaignInfo(id string) (campaign, error) {
	var camp campaign
	err := Db.QueryRow("SELECT `id`, `interface_id`, `name`, `subject`, `from`, `from_name`, `body`, `start_time`, `end_time` FROM `campaign` WHERE id=?", id).Scan(
		&camp.Id,
		&camp.IfaceId,
		&camp.Name,
		&camp.Subject,
		&camp.From,
		&camp.FromName,
		&camp.Message,
		&camp.StartTime,
		&camp.EndTime,
	)

	return camp, err
}

func getIfaces() map[string]iFace {
	var id, name, iface, host, stream, delay string
	ifaces := make(map[string]iFace)
	query, err := Db.Query("SELECT `id`, `name`, `iface`, `host`, `stream`, `delay` FROM interface")
	checkErr(err)
	defer query.Close()
	for query.Next() {
		err = query.Scan(&id, &name, &iface, &host, &stream, &delay)
		checkErr(err)
		ifaces[id] = iFace{
			Id:     id,
			Name:   name,
			Iface:  iface,
			Host:   host,
			Stream: stream,
			Delay:  delay,
		}
	}
	return ifaces
}

func updateCampaignInfo(camp campaign) campaign {

	// Fix TinyMce bug replace & to &amp; in url
	r := regexp.MustCompile(`href=["'](.*?)["']`)
	camp.Message = r.ReplaceAllStringFunc(camp.Message, func(str string) string {
		return strings.Replace(str, "&amp;", "&", -1)
	})
	_, err := Db.Query("UPDATE campaign SET `interface_id`=?, `name`=?, `subject`=?, `from`=?, `from_name`=?, `body`=?, `start_time`=?, `end_time`=? WHERE id=?",
		camp.IfaceId,
		camp.Name,
		camp.Subject,
		camp.From,
		camp.FromName,
		camp.Message,
		camp.StartTime,
		camp.EndTime,
		camp.Id,
	)
	checkErr(err)

	res, _ := getCampaignInfo(camp.Id)
	return res
}

func getRecipients(campId string, from string, limit string) map[string]recipient {
	var key, email, name string
	recipients := make(map[string]recipient)
	query, err := Db.Query("SELECT `id`, `email`, `name` FROM `recipient` WHERE `campaign_id`=? ORDER BY `id` LIMIT ?,?", campId, from, limit)
	checkErr(err)
	defer query.Close()
	for query.Next() {
		err = query.Scan(&key, &email, &name)
		checkErr(err)
		recipients[key] = recipient{Id: key, CampaignId: campId, Email: email, Name: name}
	}
	return recipients
}

func getRecipient(id string) recipient {
	var campaignId, email, name string
	err := Db.QueryRow("SELECT `campaign_id`, `email`, `name` FROM `recipient` WHERE `id`=?", id).Scan(&campaignId, &email, &name)
	checkErr(err)
	return recipient{Id: id, CampaignId: campaignId, Email: email, Name: name}
}

func getRecipientParam(id string) map[string]string {
	var paramKey, paramValue string
	recipient := make(map[string]string)
	param, err := Db.Query("SELECT `key`, `value` FROM parameter WHERE recipient_id=?", id)
	checkErr(err)
	defer param.Close()
	for param.Next() {
		err = param.Scan(&paramKey, &paramValue)
		checkErr(err)
		recipient[string(paramKey)] = string(paramValue)
	}

	return recipient
}

