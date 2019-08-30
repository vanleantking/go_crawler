package utils

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/tebeka/selenium"
	"gopkg.in/mgo.v2/bson"
)

const (
	LinkRexp = `\b(https?)`
)

// get list of comment from box_comment
func GetAllDetailCmts(detailDriver selenium.WebDriver,
	linkCrwl LinkCrwl,
	vnexUsersC *mongo.Collection) []DetailComment {
	var allComments = make(map[string]DetailComment)

	commentItemE, er := detailDriver.FindElements(
		selenium.ByCSSSelector, ".comment_item")
	if er != nil {
		log.Println("eror on get comment_item value, ", er.Error(), linkCrwl.Link)
		return []DetailComment{}
	}

	// get text full comment
	// insert list of comment and user info
	for indexCmtItem := 0; indexCmtItem < len(commentItemE); indexCmtItem++ {
		// click txt_view_more comment
		countE := 0
		var cmtItem = commentItemE[indexCmtItem]
		for {
			viewMoreRepE, er := cmtItem.FindElements(
				selenium.ByCSSSelector, ".txt_view_more a.view_all_reply")
			fmt.Println("Error txt_view_more, ", er, len(viewMoreRepE))
			if len(viewMoreRepE) == 0 {
				break
			}

			for idxViewMore := 0; idxViewMore < len(viewMoreRepE); idxViewMore++ {
				// click view more
				err := viewMoreRepE[idxViewMore].Click()
				if err != nil {
					countE++
					break
				}
				href, _ := viewMoreRepE[idxViewMore].Text()
				fmt.Println("txt view more click err, ", err, len(viewMoreRepE), href, linkCrwl.Link)
				time.Sleep(1000 * time.Millisecond)

			}
			fmt.Println("-----------------------error count, ", len(viewMoreRepE))
			if countE > 0 {
				break
			}
		}

		// click icon_show_full_comment
		for {
			viewFullCmtsE, er := cmtItem.FindElements(
				selenium.ByCSSSelector,
				".content_less .icon_show_full_comment")
			fmt.Println("Error icon_show_full_comment, ", er, len(viewFullCmtsE))
			if len(viewFullCmtsE) == 0 {
				break
			}

			for idxViewFull := 0; idxViewFull < len(viewFullCmtsE); idxViewFull++ {
				// click view more
				err := viewFullCmtsE[idxViewFull].Click()
				fmt.Println("Error icon show full comment, ", err)
				time.Sleep(2000 * time.Millisecond)
			}
			break
		}

		// check length sub_comment length class:
		// == 0 => ((full_content, user_status) | (content_more, user_status))
		// (> 0) => ((full_content, user_status) | (content_more, user_status)),
		// (sub_comment[sub_comment_item(full_content, user_status)])

		// check length .sub_comment
		subCmtE, _ := cmtItem.FindElements(
			selenium.ByCSSSelector, ".sub_comment_item")
		fmt.Println("len sub commenttttttttttttttttttttttttttt, ", len(subCmtE))
		if len(subCmtE) == 0 {
			detailCmt, er := GetDetailCmt(cmtItem, linkCrwl)
			if er != nil || detailCmt.Content == "" {
				continue
			}
			// check detail comment not exist in map comments <= add
			if _, ok := allComments[detailCmt.Content]; !ok {
				allComments[detailCmt.Content] = detailCmt
			}
		} else {
			// get first comment
			firstComment, er := GetDetailCmt(cmtItem, linkCrwl)
			if er != nil || firstComment.Content == "" {
				continue
			}
			// check detail comment not exist in map comments <= add
			if _, ok := allComments[firstComment.Content]; !ok {
				allComments[firstComment.Content] = firstComment
			}

			// get list of reply comment on sub_comment detail
			for i := 0; i < len(subCmtE); i++ {
				replyComment, er := GetDetailCmt(subCmtE[i], linkCrwl)
				if er != nil || replyComment.Content == "" {
					continue
				}
				// check detail comment not exist in map comments <= add
				if _, ok := allComments[replyComment.Content]; !ok {
					allComments[replyComment.Content] = replyComment
				}
			}
		}
	}

	detailComments := []DetailComment{}
	for _, detail := range allComments {

		linkPattern := regexp.MustCompile(LinkRexp)
		// only insert user if can  match link profile pattern
		if linkPattern.MatchString(detail.ProfileLink) {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			count, er := vnexUsersC.Count(ctx,
				bson.M{"profile_link": detail.ProfileLink})

			if er != nil {
				log.Println("Error, can not count profile from vnexpress_users, ", er.Error())
				continue
			}

			// only insert if not exist
			if count == int64(0) {
				user := VNExUser{
					ID:          primitive.NewObjectID(),
					ProfileLink: detail.ProfileLink,
					Status:      0}

				ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
				_, er := vnexUsersC.InsertOne(
					ctx,
					&user)
				if er != nil {
					log.Println("Error, can not insert user, ", user, er.Error())
				}
			}
		}
		detailComments = append(detailComments, detail)
	}
	return detailComments
}

// get detail comment {user_profile, comment}
func GetDetailCmt(cmtItem selenium.WebElement,
	linkCrwl LinkCrwl) (DetailComment, error) {

	detailComment := DetailComment{}
	// get user_info from comment element
	var userInfoE selenium.WebElement
	var er error

	// get user-info from .nickname
	userInfoE, er = cmtItem.FindElement(
		selenium.ByCSSSelector, ".nickname")
	if er != nil {

		// get user-info from .avata_coment
		userInfoE, er = cmtItem.FindElement(
			selenium.ByCSSSelector, ".avata_coment")
		if er != nil {
			log.Println("eror on get user profile link value, ", linkCrwl.Link, er.Error())
			return detailComment, er
		}
	}

	profileLink, er := userInfoE.GetAttribute("href")
	if er != nil {
		log.Println("Error, can not get profile user, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	userName, er := userInfoE.Text()
	if er != nil {
		log.Println("Error, can not get user name, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	detailComment.ProfileLink = profileLink
	detailComment.UserName = userName

	// click view full_content comment
	var fullCmtText = ""
	cmtsE, er := cmtItem.FindElement(
		selenium.ByCSSSelector, "p.full_content")
	if er == nil {
		fullCmtText, er = cmtsE.Text()
		if er != nil {
			log.Println("eror on get full_content comment value, ", linkCrwl, er.Error())
			return detailComment, er
		}
	} else {
		// click view content_more
		cmtmoresE, er := cmtItem.FindElement(
			selenium.ByCSSSelector,
			".content_more")
		if er != nil {
			log.Println("eror on get content_more view value, ", er.Error())
			return detailComment, er
		}

		fullCmtText, er = cmtmoresE.Text()
		if er != nil {
			return detailComment, er
		}
	}
	if fullCmtText == "" {
		return detailComment, er
	}
	detailComment.Content = fullCmtText
	return detailComment, nil
}
