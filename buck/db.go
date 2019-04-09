package buck

import "github.com/strongjz/slack-bucks/database"

func (b *Buck) updateDB(g database.Gift) error {

	logger.Print("[INFO] updateDB")

	err := b.db.WriteGift(&g)
	if err != nil {

		return err
	}

	return nil
}
